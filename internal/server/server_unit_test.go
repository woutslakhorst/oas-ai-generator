package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/woutslakhorst/oas-ai-generator/internal/db"
	"github.com/woutslakhorst/oas-ai-generator/internal/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	os.Setenv("DATABASE_PATH", ":memory:")
	database := db.New()
	t.Cleanup(func() { database.Close() })
	return database
}

func loadJSON(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestServer_PetHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	s := &Server{db: db}

	// addPet
	body := loadJSON(t, "../../testdata/pet.json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/pet", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	s.addPet(c)
	if w.Code != http.StatusOK {
		t.Fatalf("addPet status %d", w.Code)
	}
	var p models.Pet
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.ID == 0 {
		t.Fatal("expected id assigned")
	}

	// getPetByID
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "petId", Value: strconv.Itoa(int(p.ID))}}
	c.Request = httptest.NewRequest(http.MethodGet, "/pet/"+strconv.Itoa(int(p.ID)), nil)
	s.getPetByID(c)
	if w.Code != http.StatusOK {
		t.Fatalf("getPetByID status %d", w.Code)
	}

	// updatePet
	p.Name = "updated"
	upd, _ := json.Marshal(p)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/pet", bytes.NewReader(upd))
	c.Request.Header.Set("Content-Type", "application/json")
	s.updatePet(c)
	if w.Code != http.StatusOK {
		t.Fatalf("updatePet status %d", w.Code)
	}

	// updatePetWithForm
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/pet/"+strconv.Itoa(int(p.ID))+"?name=form&status=pending", nil)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "petId", Value: strconv.Itoa(int(p.ID))}}
	s.updatePetWithForm(c)
	if w.Code != http.StatusOK {
		t.Fatalf("updatePetWithForm status %d", w.Code)
	}

	// findPetsByStatus
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/pet/findByStatus?status=pending", nil)
	s.findPetsByStatus(c)
	if w.Code != http.StatusOK {
		t.Fatalf("findPetsByStatus status %d", w.Code)
	}

	// getInventory
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/store/inventory", nil)
	s.getInventory(c)
	if w.Code != http.StatusOK {
		t.Fatalf("getInventory status %d", w.Code)
	}

	// placeOrder
	orderBody := loadJSON(t, "../../testdata/order.json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/store/order", bytes.NewReader(orderBody))
	c.Request.Header.Set("Content-Type", "application/json")
	s.placeOrder(c)
	if w.Code != http.StatusOK {
		t.Fatalf("placeOrder status %d", w.Code)
	}
	var o models.Order
	json.Unmarshal(w.Body.Bytes(), &o)

	// getOrderByID
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "orderId", Value: strconv.Itoa(int(o.ID))}}
	c.Request = httptest.NewRequest(http.MethodGet, "/store/order/"+strconv.Itoa(int(o.ID)), nil)
	s.getOrderByID(c)
	if w.Code != http.StatusOK {
		t.Fatalf("getOrderByID status %d", w.Code)
	}

	// deleteOrder
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "orderId", Value: strconv.Itoa(int(o.ID))}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/store/order/"+strconv.Itoa(int(o.ID)), nil)
	s.deleteOrder(c)
	if w.Code != http.StatusOK {
		t.Fatalf("deleteOrder status %d", w.Code)
	}

	// createUser
	userBody := loadJSON(t, "../../testdata/user.json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(userBody))
	c.Request.Header.Set("Content-Type", "application/json")
	s.createUser(c)
	if w.Code != http.StatusOK {
		t.Fatalf("createUser status %d", w.Code)
	}
	var u models.User
	json.Unmarshal(w.Body.Bytes(), &u)

	// createUsersWithListInput
	listBody := `[` + string(userBody) + `]`
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/user/createWithList", bytes.NewReader([]byte(listBody)))
	c.Request.Header.Set("Content-Type", "application/json")
	s.createUsersWithListInput(c)
	if w.Code != http.StatusOK {
		t.Fatalf("createUsersWithListInput status %d", w.Code)
	}

	// loginUser
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	q := "?username=" + u.Username + "&password=" + u.Password
	c.Request = httptest.NewRequest(http.MethodGet, "/user/login"+q, nil)
	s.loginUser(c)
	if w.Code != http.StatusOK {
		t.Fatalf("loginUser status %d", w.Code)
	}

	// getUserByName
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "username", Value: u.Username}}
	c.Request = httptest.NewRequest(http.MethodGet, "/user/"+u.Username, nil)
	s.getUserByName(c)
	if w.Code != http.StatusOK {
		t.Fatalf("getUserByName status %d", w.Code)
	}

	// updateUser
	u.FirstName = "Updated"
	updUser, _ := json.Marshal(u)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "username", Value: u.Username}}
	c.Request = httptest.NewRequest(http.MethodPut, "/user/"+u.Username, bytes.NewReader(updUser))
	c.Request.Header.Set("Content-Type", "application/json")
	s.updateUser(c)
	if w.Code != http.StatusOK {
		t.Fatalf("updateUser status %d", w.Code)
	}

	// logoutUser
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/user/logout", nil)
	s.logoutUser(c)
	if w.Code != http.StatusOK {
		t.Fatalf("logoutUser status %d", w.Code)
	}

	// deleteUser
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "username", Value: u.Username}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/user/"+u.Username, nil)
	s.deleteUser(c)
	if w.Code != http.StatusOK {
		t.Fatalf("deleteUser status %d", w.Code)
	}

	// deletePet
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "petId", Value: strconv.Itoa(int(p.ID))}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/pet/"+strconv.Itoa(int(p.ID)), nil)
	s.deletePet(c)
	if w.Code != http.StatusOK {
		t.Fatalf("deletePet status %d", w.Code)
	}
}

func TestServer_uploadFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	s := &Server{db: db}

	// first add pet to have ID
	body := loadJSON(t, "../../testdata/pet.json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/pet", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	s.addPet(c)
	var p models.Pet
	json.Unmarshal(w.Body.Bytes(), &p)

	// upload file
	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)
	fw, _ := mw.CreateFormFile("file", "test.txt")
	fw.Write([]byte("test"))
	mw.Close()

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/pet/"+strconv.Itoa(int(p.ID))+"/uploadImage", buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "petId", Value: strconv.Itoa(int(p.ID))}}
	s.uploadFile(c)
	if w.Code != http.StatusOK {
		t.Fatalf("uploadFile status %d", w.Code)
	}
}
