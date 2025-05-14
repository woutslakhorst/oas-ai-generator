package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/woutslakhorst/oas-ai-generator/internal/models"
)

func setupRouter(t *testing.T) (*gin.Engine, *sql.DB) {
	db := setupTestDB(t)
	return New(db), db
}

func TestIntegration_PetOrderUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupRouter(t)
	server := httptest.NewServer(router)
	defer server.Close()

	// create pet
	petBody := loadJSON(t, "testdata/pet.json")
	resp, err := http.Post(server.URL+"/pet", "application/json", bytes.NewReader(petBody))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("add pet %d", resp.StatusCode)
	}
	var p models.Pet
	json.NewDecoder(resp.Body).Decode(&p)
	resp.Body.Close()

	// update pet
	p.Name = "updated"
	buf, _ := json.Marshal(p)
	req, _ := http.NewRequest(http.MethodPut, server.URL+"/pet", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// update with form
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/pet/"+strconv.Itoa(int(p.ID))+"?name=form&status=sold", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// get pet
	resp, err = http.Get(server.URL + "/pet/" + strconv.Itoa(int(p.ID)))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// find by status
	http.Get(server.URL + "/pet/findByStatus?status=sold")

	// insert tag for findByTags
	db.Exec(`INSERT INTO tags (name) VALUES ('tag1')`)
	db.Exec(`INSERT INTO pet_tags (pet_id, tag_id) VALUES (?, 1)`, p.ID)
	http.Get(server.URL + "/pet/findByTags?tags=tag1")

	// upload file
	buf2 := new(bytes.Buffer)
	mw := multipart.NewWriter(buf2)
	fw, _ := mw.CreateFormFile("file", "test.txt")
	fw.Write([]byte("data"))
	mw.Close()
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/pet/"+strconv.Itoa(int(p.ID))+"/uploadImage", buf2)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	http.DefaultClient.Do(req)

	// get inventory
	http.Get(server.URL + "/store/inventory")

	// place order
	orderBody := loadJSON(t, "testdata/order.json")
	resp, err = http.Post(server.URL+"/store/order", "application/json", bytes.NewReader(orderBody))
	if err != nil {
		t.Fatal(err)
	}
	var o models.Order
	json.NewDecoder(resp.Body).Decode(&o)
	resp.Body.Close()

	// get order
	resp, err = http.Get(server.URL + "/store/order/" + strconv.Itoa(int(o.ID)))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// create user
	userBody := loadJSON(t, "testdata/user.json")
	resp, err = http.Post(server.URL+"/user", "application/json", bytes.NewReader(userBody))
	if err != nil {
		t.Fatal(err)
	}
	var u models.User
	json.NewDecoder(resp.Body).Decode(&u)
	resp.Body.Close()

	// create user list
	listBody := `[` + string(userBody) + `]`
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/user/createWithList", bytes.NewReader([]byte(listBody)))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)

	// login user
	resp, err = http.Get(server.URL + "/user/login?username=" + u.Username + "&password=" + u.Password)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// get user by name
	http.Get(server.URL + "/user/" + u.Username)

	// update user
	u.FirstName = "Updated"
	up, _ := json.Marshal(u)
	req, _ = http.NewRequest(http.MethodPut, server.URL+"/user/"+u.Username, bytes.NewReader(up))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)

	// logout
	http.Get(server.URL + "/user/logout")

	// delete order
	req, _ = http.NewRequest(http.MethodDelete, server.URL+"/store/order/"+strconv.Itoa(int(o.ID)), nil)
	http.DefaultClient.Do(req)

	// delete user
	req, _ = http.NewRequest(http.MethodDelete, server.URL+"/user/"+u.Username, nil)
	http.DefaultClient.Do(req)

	// delete pet
	req, _ = http.NewRequest(http.MethodDelete, server.URL+"/pet/"+strconv.Itoa(int(p.ID)), nil)
	http.DefaultClient.Do(req)
}
