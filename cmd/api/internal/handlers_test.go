package internal_test

import (
	"bookstore-api/cmd/api/internal"
	"bytes"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
)

func TestNewHandlers_InvalidDatabaseConnection(t *testing.T) {
	// Given
	var db *gorm.DB

	// When
	h, err := internal.NewHandlers(db, nil, nil)

	// Then
	require.Nil(t, h)
	require.EqualError(t, err, "please provide a database connection")
}

func TestNewHandlers_ValidConnection(t *testing.T) {
	// Given
	gormDb, _, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	// When
	h, err := internal.NewHandlers(gormDb, nil, nil)

	// Then
	require.NotNil(t, h)
	require.Nil(t, err)
}

func mockDatabaseConnection(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func() error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	closeFunc := db.Close

	mock.ExpectQuery("SELECT VERSION()").
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.0"))

	dialector := mysql.New(mysql.Config{
		DSN:        "sqlmock_db_0",
		DriverName: "mysql",
		Conn:       db,
	})

	gormDb, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return gormDb, mock, closeFunc
}

func TestHandlers_Index(t *testing.T) {
	// Given
	gormDb, _, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// When
	h.Index(c)

	// Then
	//que me responde el handler?
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, `{"message":"holis"}`, w.Body.String())
}

func TestHandlers_RegisterInvalidRequestBody(t *testing.T) {
	// Given
	gormDb, _, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{INVALID}`
	request, _ := http.NewRequest(http.MethodPost, "/register", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	// When
	h.RegisterHandler(c)

	// Then
	//que me responde el handler?
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"invalid character 'I' looking for beginning of object key string"}`, w.Body.String())
}

func TestHandlers_RegisterInvalidPasswordVerification(t *testing.T) {
	// Given
	gormDb, _, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"Email": "test@test",
		"Password": "123456",
		"Pass_confirmation": "LALALLA",
		"Name": "keseyo",
		"Last_name": "tuvieja"
	}`
	request, _ := http.NewRequest(http.MethodPost, "/register", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	// When
	h.RegisterHandler(c)

	// Then
	//que me responde el handler?
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"password_confirmation_failed"}`, w.Body.String())
}

func TestHandlers_RegisterFailedToHashPassword(t *testing.T) {
	// Given
	gormDb, _, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, func(pass string) (string, error) {
		return "", errors.New("failed to hash password")
	}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"Email": "test@test",
		"Password": "123456",
		"Pass_confirmation": "123456",
		"Name": "keseyo",
		"Last_name": "tuvieja"
	}`
	request, _ := http.NewRequest(http.MethodPost, "/register", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	// When
	h.RegisterHandler(c)

	// Then
	//que me responde el handler?
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Something went wrong with the hash"}`, w.Body.String())
}

func TestHandlers_RegisterFailedCreateUser(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, func(pass string) (string, error) {
		return "algo", nil
	}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"Email": "test@test",
		"Password": "123456",
		"Pass_confirmation": "123456",
		"Name": "keseyo",
		"Last_name": "tuvieja"
	}`
	request, _ := http.NewRequest(http.MethodPost, "/register", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users` (`email`,`password`,`name`,`last_name`,`role`) VALUES (?,?,?,?,?)")).
		WithArgs("test@test", "algo", "keseyo", "tuvieja", "user").
		WillReturnError(errors.New("unknown error"))

	mock.ExpectRollback()

	// When
	h.RegisterHandler(c)

	// Then
	//que me responde el handler?
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Something went wrong"}`, w.Body.String())
}

func TestHandlers_RegisterSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, func(pass string) (string, error) {
		return "algo", nil
	}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"Email": "test@test",
		"Password": "123456",
		"Pass_confirmation": "123456",
		"Name": "keseyo",
		"Last_name": "tuvieja"
	}`
	request, _ := http.NewRequest(http.MethodPost, "/register", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users` (`email`,`password`,`name`,`last_name`,`role`) VALUES (?,?,?,?,?)")).
		WithArgs("test@test", "algo", "keseyo", "tuvieja", "user").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	// When
	h.RegisterHandler(c)

	// Then
	//que me responde el handler?
	require.Equal(t, http.StatusCreated, w.Code)
	require.Equal(t, `"test@test"`, w.Body.String())
}

func TestHandlers_LoginInvalidRequestBody(t *testing.T) {
	// Given
	gormDb, _, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{INVALID}`
	request, _ := http.NewRequest(http.MethodPost, "/login", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	h.LoginHandler(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"invalid character 'I' looking for beginning of object key string"}`, w.Body.String())
}

func TestHandlers_LoginInvalidSearchQuery(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"Email": "test@test", "Password": "123456"}`
	request, _ := http.NewRequest(http.MethodPost, "/login", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? LIMIT 1")).
		WithArgs("test@test").
		WillReturnError(errors.New("unknown error"))

	// When
	h.LoginHandler(c)

	// Then
	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Equal(t, `{"error":"Something went wrong with the Email"}`, w.Body.String())
}

func TestHandlers_LoginInvalidPassword(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, func(pass string, input string) error {
		return errors.New("invalid password")
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"Email": "test@test", "Password": "123456"}`
	request, _ := http.NewRequest(http.MethodPost, "/login", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? LIMIT 1")).
		WithArgs("test@test").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// When
	h.LoginHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Something went wrong with the password"}`, w.Body.String())
}

func TestHandlers_LoginFailedToGenerateToken(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, func(pass string, input string) error {
		return nil
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"Email": "test@test", "Password": "123456"}`
	request, _ := http.NewRequest(http.MethodPost, "/login", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? LIMIT 1")).
		WithArgs("test@test").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	//This ENV variables are required for the GenerateToken to work, if not defined it fails
	//os.Setenv("TOKEN_HOUR_LIFESPAN", "100")
	//os.Setenv("TOKEN_KEY", "LALALA")

	// When
	h.LoginHandler(c)

	// Then
	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Equal(t, `{"error":"Something went wrong with the token: strconv.Atoi: parsing \"\": invalid syntax"}`, w.Body.String())
}

func TestHandlers_LoginSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, func(pass string, input string) error {
		return nil
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"Email": "test@test", "Password": "123456"}`
	request, _ := http.NewRequest(http.MethodPost, "/login", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? LIMIT 1")).
		WithArgs("test@test").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	os.Setenv("TOKEN_HOUR_LIFESPAN", "100")
	os.Setenv("TOKEN_KEY", "LALALA")

	// When
	h.LoginHandler(c)

	// Then
	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Body.String())
}

func TestHandlers_GetBooksByIdInvalidId(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "assad"}}

	request, _ := http.NewRequest(http.MethodGet, "/books/:ID", nil)
	c.Request = request

	// When
	h.GetBookByIDHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Something went wrong with the ID"}`, w.Body.String())
}

func TestHandlers_GetBooksByIdFailedQuery(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "1"}}

	request, _ := http.NewRequest(http.MethodGet, "/books/:ID", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE ID = ?")).
		WithArgs("").
		WillReturnError(errors.New("unknown error"))

	// When
	h.GetBookByIDHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Something went wrong with the ID"}`, w.Body.String())
}

func TestHandlers_GetBooksByIdSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "1"}}

	request, _ := http.NewRequest(http.MethodGet, "/books/:ID", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE ID = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "category", "price", "Description", "Language", "Cover", "Editorial", "Year", "Pages"}).
			AddRow(1, "sarasa", "pepa", "fiction", 123, "pato", "español", "blanda", "lala", 2004, 123))

	// When
	h.GetBookByIDHandler(c)

	// Then
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, `{"ID":1,"Title":"sarasa","Author":"pepa","Category":"fiction","Price":123,"Description":"pato","Language":"español","Cover":"blanda","Editorial":"lala","Year":2004,"Pages":123}`, w.Body.String())
}

func TestHandlers_GetBooksByCategoryInvalidPagesPagination(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "page", Value: "nan"}}

	request, _ := http.NewRequest(http.MethodGet, "/books/category/category/fiction?page=nan", nil)
	c.Request = request

	// When
	h.GetBooksByCategoryHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Something went wrong with the pagination"}`, w.Body.String())
}

func TestHandlers_GetBooksByCategoryInvalidLimitPagination(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "limit", Value: "nan"}}

	request, _ := http.NewRequest(http.MethodGet, "/books/category/category/fiction?limit=nan", nil)
	c.Request = request

	// When
	h.GetBooksByCategoryHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Something went wrong"}`, w.Body.String())
}

func TestHandlers_GetBooksByCategoryFailedQuery(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "category", Value: "fiction"}}

	request, _ := http.NewRequest(http.MethodGet, "/books/category/:category", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE category = ?")).
		WithArgs("").
		WillReturnError(errors.New("unknown error"))

	// When
	h.GetBooksByCategoryHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Something went wrong while fetching the books"}`, w.Body.String())
}

func TestHandlers_GetBooksByCategorySuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "category", Value: "fiction"}}

	request, _ := http.NewRequest(http.MethodGet, "/books/category/:category", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE category = ?")).
		WithArgs("fiction").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "category", "price", "Description", "Language", "Cover", "Editorial", "Year", "Pages"}).
			AddRow(1, "sarasa", "pepa", "fiction", 123, "pato", "español", "blanda", "lala", 2004, 123))

	// When
	h.GetBooksByCategoryHandler(c)

	// Then
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, `[{"ID":1,"Title":"sarasa","Author":"pepa","Category":"fiction","Price":123,"Description":"pato","Language":"español","Cover":"blanda","Editorial":"lala","Year":2004,"Pages":123}]`, w.Body.String())
}

func TestHandlers_GetBooksByAuthorFailedQuery(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "author", Value: "tolkien"}}

	request, _ := http.NewRequest(http.MethodGet, "/books/author/:author", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE author = ?")).
		WithArgs("").
		WillReturnError(errors.New("unknown error"))

	// When
	h.GetBooksByAuthorHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"There is not book for the Author"}`, w.Body.String())
}

func TestHandlers_GetBooksByAuthorSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "author", Value: "sarasa"}}

	request, _ := http.NewRequest(http.MethodGet, "/books/author/:author", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE author = ?")).
		WithArgs("sarasa").
		WithArgs("sarasa").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "category", "price", "Description", "Language", "Cover", "Editorial", "Year", "Pages"}).
			AddRow(1, "sarasa", "pepa", "fiction", 123, "pato", "español", "blanda", "lala", 2004, 123))

	// When
	h.GetBooksByAuthorHandler(c)

	// Then
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, `[{"ID":1,"Title":"sarasa","Author":"pepa","Category":"fiction","Price":123,"Description":"pato","Language":"español","Cover":"blanda","Editorial":"lala","Year":2004,"Pages":123}]`, w.Body.String())
}

func TestHandlers_SearchBookInvalidQuery(t *testing.T) {
	// Given
	gormDb, _, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	request, _ := http.NewRequest(http.MethodGet, "/books/search?query=", nil)
	c.Request = request

	// When
	h.SearchBookHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `"error": "Invalid query"`, w.Body.String())
}

func TestHandlers_SearchBookFailedQuery(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	//c.Request.URL.RawQuery = ""

	request, _ := http.NewRequest(http.MethodGet, "/books/search?query=sarasa", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE title LIKE ? OR author LIKE ?")).
		WithArgs("").
		WillReturnError(errors.New("unknown error"))

	// When
	h.SearchBookHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Book or Author not found"}`, w.Body.String())
}

func TestHandlers_SearchBookSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	request, _ := http.NewRequest(http.MethodGet, "/books/search?query=sarasa", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE title LIKE ? OR author LIKE ?")).
		WithArgs("%sarasa%", "%sarasa%").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "category", "price", "Description", "Language", "Cover", "Editorial", "Year", "Pages"}).
			AddRow(1, "sarasa", "pepa", "fiction", 123, "pato", "español", "blanda", "lala", 2004, 123))

	// When
	h.SearchBookHandler(c)

	// Then
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, `[{"ID":1,"Title":"sarasa","Author":"pepa","Category":"fiction","Price":123,"Description":"pato","Language":"español","Cover":"blanda","Editorial":"lala","Year":2004,"Pages":123}]`, w.Body.String())
}

func TestHandlers_AddNewBookInvalidRequestBody(t *testing.T) {
	// Given
	gormDb, _, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{INVALID}`
	request, _ := http.NewRequest(http.MethodPost, "/admin/books", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	// When
	h.AddNewBookHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"invalid character 'I' looking for beginning of object key string"}`, w.Body.String())
}

func TestHandlers_AddNewBookFailedCreation(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{""ID":1,"Title":"sarasa","Author":"pepa","Category":"fiction","Price":123,"Description":"pato","Language":"español","Cover":"blanda","Editorial":"lala","Year":2004,"Pages":123"}`
	request, _ := http.NewRequest(http.MethodPost, "/admin/books", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `products` (`Title`, `Author`, `Category`, `Price`, `Description`, `Language`, `Language`, `Cover`, `Editorial`, `Year`, `Pages`) VALUES (?,?,?,?,?,?,?,?,?,?,?)")).
		WithArgs("sarasa", "pepa", "fiction", 123, "pato", "español", "blanda", "lala", 2004, 123).
		WillReturnError(errors.New("unknown error"))

	mock.ExpectCommit()
	// When
	h.AddNewBookHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"invalid character 'I' after object key"}`, w.Body.String())
}

func TestHandlers_AddNewBookSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"Title":"sarasa",
		"Author":"pepa",
		"Category":"fiction",
		"Price":123.50,
		"Description":"pato",
		"Language":"español",
		"Cover":"blanda",
		"Editorial":"lala",
		"Year":2004,
		"Pages":123
	}`
	request, _ := http.NewRequest(http.MethodPost, "/admin/books", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `products` (`title`, `author`, `category`, `price`, `description`, `language`, `cover`, `editorial`, `year`, `pages`) VALUES (?,?,?,?,?,?,?,?,?,?)")).
		WithArgs("sarasa", "pepa", "fiction", 123.50, "pato", "español", "blanda", "lala", 2004, 123).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	// When
	h.AddNewBookHandler(c)

	// Then
	require.Equal(t, http.StatusCreated, w.Code)
	require.Equal(t, `{"error":"invalid character 'I' after object key"}`, w.Body.String())
}

func TestHandlers_DeleteBookInvalidId(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "assad"}}

	request, _ := http.NewRequest(http.MethodDelete, "/admin/books/:ID", nil)
	c.Request = request

	// When
	h.DeleteBookHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Something went wrong with the ID"}`, w.Body.String())
}

func TestHandlers_DeleteBookFailedQuery(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "1"}}

	request, _ := http.NewRequest(http.MethodDelete, "/books/:ID", nil)
	c.Request = request

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `products` WHERE ID = ?")).
		WithArgs(1).
		WillReturnError(errors.New("unknown error"))

	mock.ExpectCommit()

	// When
	h.DeleteBookHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Book not deleted"}`, w.Body.String())
}

func TestHandlers_DeleteBookSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "1"}}

	request, _ := http.NewRequest(http.MethodDelete, "/books/:ID", nil)
	c.Request = request

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `products` WHERE ID = ?")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	// When
	h.DeleteBookHandler(c)

	// Then
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, `{"ID":0,"Title":"","Author":"","Category":"","Price":0,"Description":"","Language":"","Cover":"","Editorial":"","Year":0,"Pages":0}`, w.Body.String())
}

func TestHandlers_UpdateBookInvalidId(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "assad"}}

	request, _ := http.NewRequest(http.MethodPatch, "/admin/books/:ID", nil)
	c.Request = request

	// When
	h.UpdateBookHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Something went wrong with the ID"}`, w.Body.String())
}

func TestHandlers_UpdateBookFailedSearch(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "1"}}

	request, _ := http.NewRequest(http.MethodPatch, "/admin/books/:ID", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE `products`.`id` = ?")).
		WithArgs(1).
		WillReturnError(errors.New("unknown error"))
	// When
	h.UpdateBookHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Book not found"}`, w.Body.String())
}

func TestHandlers_UpdateBookInvalidBody(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "1"}}

	body := `{INVALID}`
	request, _ := http.NewRequest(http.MethodPatch, "/admin/books/:ID", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE `products`.`id` = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "category", "price", "Description", "Language", "Cover", "Editorial", "Year", "Pages"}).
			AddRow(1, "sarasa", "pepa", "fiction", 123, "pato", "español", "blanda", "lala", 2004, 123))

	// When
	h.UpdateBookHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"invalid character 'I' looking for beginning of object key string"}`, w.Body.String())
}

func TestHandlers_UpdateBookFailedUpdate(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "1"}}

	body := `{"Pages":1234}`
	request, _ := http.NewRequest(http.MethodPatch, "/admin/books/:ID", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE `products`.`id` = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "category", "price", "Description", "Language", "Cover", "Editorial", "Year", "Pages"}).
			AddRow(1, "sarasa", "pepa", "fiction", 123, "pato", "español", "blanda", "lala", 2004, 123))

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE `products` SET `id`=?,`title`=?,`author`=?,`category`=?,`price`=?,`description`=?,`language`=?,`cover`=?,`editorial`=?,`year`=?,`pages`=? WHERE `id` = ?")).
		WithArgs(1, "sarasa", "pepa", "fiction", 123, "pato", "español", "blanda", "lala", 2004, 1234, 1).
		WillReturnError(errors.New("unknown error"))

	mock.ExpectCommit()
	// When
	h.UpdateBookHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Book not updated"}`, w.Body.String())
}

func TestHandlers_UpdateBookSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "1"}}

	body := `{"Pages":1234}`
	request, _ := http.NewRequest(http.MethodPatch, "/admin/books/:ID", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE `products`.`id` = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "category", "price", "Description", "Language", "Cover", "Editorial", "Year", "Pages"}).
			AddRow(1, "sarasa", "pepa", "fiction", 123.00, "pato", "español", "blanda", "lala", 2004, 123))

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE `products` SET `id`=?,`title`=?,`author`=?,`category`=?,`price`=?,`description`=?,`language`=?,`cover`=?,`editorial`=?,`year`=?,`pages`=? WHERE `id` = ?")).
		WithArgs(1, "sarasa", "pepa", "fiction", 123.00, "pato", "español", "blanda", "lala", 2004, 1234, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()
	// When
	h.UpdateBookHandler(c)

	// Then
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, `{"ID":1,"Title":"sarasa","Author":"pepa","Category":"fiction","Price":123,"Description":"pato","Language":"español","Cover":"blanda","Editorial":"lala","Year":2004,"Pages":1234}`, w.Body.String())
}

func TestHandlers_GetAddressIdNotFound(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("", nil)

	request, _ := http.NewRequest(http.MethodGet, "/user/address", nil)
	c.Request = request

	// When
	h.GetAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"User ID not found"}`, w.Body.String())
}

func TestHandlers_GetAddressInvalidId(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "hola")

	request, _ := http.NewRequest(http.MethodGet, "/user/address", nil)
	c.Request = request

	// When
	h.GetAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Invalid user ID"}`, w.Body.String())
}

func TestHandlers_GetAddressFailedQuery(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(1))

	request, _ := http.NewRequest(http.MethodGet, "/user/address", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_addresses` WHERE userID = ?")).
		WithArgs(1).
		WillReturnError(errors.New("unknown error"))

	// When
	h.GetAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"User address not found"}`, w.Body.String())
}

func TestHandlers_GetAddressSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(1))

	request, _ := http.NewRequest(http.MethodGet, "/user/address", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_addresses` WHERE userID = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "userID", "street", "number", "city", "postal_code", "province"}).
			AddRow(1, 1, "calle falsa", 123, "Springfield", 1234, "lalala"))

	// When
	h.GetAddressHandler(c)

	// Then
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, `{"ID":1,"UserID":1,"Street":"calle falsa","Number":123,"City":"Springfield","Postal_code":1234,"Province":"lalala"}`, w.Body.String())
}

func TestHandlers_AddAddressInvalidBody(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{INVALID}`
	request, _ := http.NewRequest(http.MethodPost, "/user/adress", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	// When
	h.AddAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"invalid character 'I' looking for beginning of object key string"}`, w.Body.String())
}

func TestHandlers_AddAddressIdNotFound(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("", nil)

	body := `{
		"ID":1,
		"UserID":1,
		"Street":"calle falsa",
		"Number":123,
		"City":"Springfield",
		"Postal_code":1234,
		"Province":"lalala"
		}`
	request, _ := http.NewRequest(http.MethodPost, "/user/address", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	// When
	h.AddAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"User ID not found"}`, w.Body.String())
}

func TestHandlers_AddAddressInvalidId(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "hola")

	body := `{
		"ID":1,
		"UserID":1,
		"Street":"calle falsa",
		"Number":123,
		"City":"Springfield",
		"Postal_code":1234,
		"Province":"lalala"
		}`
	request, _ := http.NewRequest(http.MethodPost, "/user/address", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	// When
	h.AddAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Invalid user ID"}`, w.Body.String())
}

func TestHandlers_AddAddressFailedQuery(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(1))

	body := `{
		"ID":1,
		"UserID":1,
		"Street":"calle falsa",
		"Number":123,
		"City":"Springfield",
		"Postal_code":1234,
		"Province":"lalala"
		}`
	request, _ := http.NewRequest(http.MethodPost, "/user/address", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_addresses` WHERE userID = ?")).
		WithArgs(1).
		WillReturnError(errors.New("unknown error"))

	// When
	h.AddAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"User address not created"}`, w.Body.String())
}

func TestHandlers_AddAddressSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(1))

	body := `{
		"UserID":1,
		"Street":"calle falsa",
		"Number":123,
		"City":"Springfield",
		"Postal_code":1234,
		"Province":"lalala"
		}`
	request, _ := http.NewRequest(http.MethodPost, "/user/address", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `user_addresses` (`user_id`,`street`,`number`,`city`,`postal_code`,`province`) VALUES (?,?,?,?,?,?)")).
		WithArgs(1, "calle falsa", 123, "Springfield", 1234, "lalala").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()
	// When
	h.AddAddressHandler(c)

	// Then
	require.Equal(t, http.StatusCreated, w.Code)
	require.Equal(t, `{"ID":1,"UserID":1,"Street":"calle falsa","Number":123,"City":"Springfield","Postal_code":1234,"Province":"lalala"}`, w.Body.String())
}

func TestHandlers_UpdateAddressIdNotFound(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("", nil)

	request, _ := http.NewRequest(http.MethodPatch, "/user/address", nil)
	c.Request = request

	// When
	h.UpdateAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"User ID not found"}`, w.Body.String())
}

func TestHandlers_UpdateAddressInvalidId(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "hola")

	request, _ := http.NewRequest(http.MethodPatch, "/user/address", nil)
	c.Request = request

	// When
	h.UpdateAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Invalid user ID"}`, w.Body.String())
}

func TestHandlers_UpdateAddressFailedSearch(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(1))

	request, _ := http.NewRequest(http.MethodPatch, "/user/address", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_addresses` WHERE userID = ?")).
		WithArgs(1).
		WillReturnError(errors.New("unknown error"))

	// When
	h.UpdateAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Address not found"}`, w.Body.String())
}

func TestHandlers_UpdateAddressInvalidJson(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(1))

	body := `{INVALID}`
	request, _ := http.NewRequest(http.MethodPatch, "/user/address", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_addresses` WHERE userID = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "userID", "street", "number", "city", "postal_code", "province"}).
			AddRow(1, 1, "calle falsa", 123, "Springfield", 1234, "lalala"))

	// When
	h.UpdateAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"invalid character 'I' looking for beginning of object key string"}`, w.Body.String())
}

func TestHandlers_UpdateAddressFailedUpdate(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(1))

	body := `{
		"UserID":1,
		"Street":"calle falsa",
		"Number":123,
		"City":"Springfield",
		"Postal_code":1234,
		"Province":"lalala"
		}`
	request, _ := http.NewRequest(http.MethodPatch, "/user/address", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_addresses` WHERE userID = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "userID", "street", "number", "city", "postal_code", "province"}).
			AddRow(1, 1, "calle falsa", 123, "Springfield", 1234, "lalala"))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `user_addresses` SET (`street`=?,`number`=?,`city`=?,`postal_code`=?,`province`=?) WHERE userID =?")).
		WithArgs("calle falsa", 123, "Springfield", 1234, "lalala").
		WillReturnError(errors.New("unknown error"))
	mock.ExpectCommit()

	// When
	h.UpdateAddressHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Book not updated"}`, w.Body.String())
}

func TestHandlers_UpdateAddressSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(1))

	body := `{
		"UserID":1,
		"Street":"calle falsa",
		"Number":123,
		"City":"Springfield",
		"Postal_code":1234,
		"Province":"lalala"
		}`
	request, _ := http.NewRequest(http.MethodPatch, "/user/address", io.NopCloser(bytes.NewBufferString(body)))
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `user_addresses` WHERE userID = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "userID", "street", "number", "city", "postal_code", "province"}).
			AddRow(1, 1, "calle falsa", 123, "Springfield", 123, "lalala"))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `user_addresses` SET `user_id`=?, `street`=?,`number`=?,`city`=?,`postal_code`=?,`province`=? WHERE `id`=?")).
		WithArgs(1, "calle falsa", 123, "Springfield", 1234, "lalala", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// When
	h.UpdateAddressHandler(c)

	// Then
	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{
		"ID":1,
		"UserID":1,
		"Street":"calle falsa",
		"Number":123,
		"City":"Springfield",
		"Postal_code":1234,
		"Province":"lalala"
	}`, w.Body.String())
}

func TestHandlers_GetOrderIdNotFound(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("", nil)

	request, _ := http.NewRequest(http.MethodGet, "/order", nil)
	c.Request = request

	// When
	h.GetOrdersHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"User ID not found"}`, w.Body.String())
}

func TestHandlers_GetOrderInvalidId(t *testing.T) {
	// Given
	h, _ := internal.NewHandlers(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "hola")

	request, _ := http.NewRequest(http.MethodGet, "/order", nil)
	c.Request = request

	// When
	h.GetOrdersHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"Invalid user ID"}`, w.Body.String())
}

func TestHandlers_GetOrderFailedSearch(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(1))

	request, _ := http.NewRequest(http.MethodGet, "/order", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `orders` WHERE `orders`.`user_id` = ? AND `orders`.`deleted_at` IS NULL LIMIT 10")).
		WithArgs(1).
		WillReturnError(errors.New("unknown error"))

	// When
	h.GetOrdersHandler(c)

	// Then
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, `{"error":"User order not found"}`, w.Body.String())
}

func TestHandlers_GetOrderSuccess(t *testing.T) {
	// Given
	gormDb, mock, closeFunc := mockDatabaseConnection(t)
	defer closeFunc()

	h, _ := internal.NewHandlers(gormDb, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(1))

	request, _ := http.NewRequest(http.MethodGet, "/order", nil)
	c.Request = request

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `orders` WHERE `orders`.`user_id` = ? AND `orders`.`deleted_at` IS NULL LIMIT 10")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "userID", "total"}).
			AddRow(1, 1, 100.00))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`id` = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "last_name", "role"}).
			AddRow(1, "pepa", "pepita", "user"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `order_details` WHERE `order_details`.`order_id` = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "total"}).
			AddRow(1, 1, 1, 2, 100.00))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `products` WHERE `products`.`id` = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "category", "price", "Description", "Language", "Cover", "Editorial", "Year", "Pages"}).
			AddRow(1, "sarasa", "pepa", "fiction", 123, "pato", "español", "blanda", "lala", 2004, 123))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `payments` WHERE `payments`.`order_id` = ? AND `payments`.`deleted_at` IS NULL")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "external_payment_id", "order_id", "user_id", "total"}).
			AddRow(1, "ext123", 1, 1, 100.0))

	// When
	h.GetOrdersHandler(c)

	// Then
	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"error":"User order not found"}`, w.Body.String())
}
