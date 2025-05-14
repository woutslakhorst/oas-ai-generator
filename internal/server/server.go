package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/woutslakhorst/oas-ai-generator/internal/models"
)

type Server struct {
	db *sql.DB
}

// New returns a server with configured routes.
func New(db *sql.DB) *gin.Engine {
	s := &Server{db: db}
	r := gin.Default()

	r.POST("/pet", s.addPet)
	r.PUT("/pet", s.updatePet)
	r.GET("/pet/findByStatus", s.findPetsByStatus)
	r.GET("/pet/findByTags", s.findPetsByTags)
	r.GET("/pet/:petId", s.getPetByID)
	r.POST("/pet/:petId", s.updatePetWithForm)
	r.DELETE("/pet/:petId", s.deletePet)
	r.POST("/pet/:petId/uploadImage", s.uploadFile)

	r.GET("/store/inventory", s.getInventory)
	r.POST("/store/order", s.placeOrder)
	r.GET("/store/order/:orderId", s.getOrderByID)
	r.DELETE("/store/order/:orderId", s.deleteOrder)

	r.POST("/user", s.createUser)
	r.POST("/user/createWithList", s.createUsersWithListInput)
	r.GET("/user/login", s.loginUser)
	r.GET("/user/logout", s.logoutUser)
	r.GET("/user/:username", s.getUserByName)
	r.PUT("/user/:username", s.updateUser)
	r.DELETE("/user/:username", s.deleteUser)

	return r
}

func (s *Server) addPet(c *gin.Context) {
	var p models.Pet
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stmt, err := s.db.Prepare(`INSERT INTO pets (name, status) VALUES (?, ?)`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	res, err := stmt.Exec(p.Name, p.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	p.ID = id
	c.JSON(http.StatusOK, p)
}

func (s *Server) updatePet(c *gin.Context) {
	var p models.Pet
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stmt, err := s.db.Prepare(`UPDATE pets SET name=?, status=? WHERE id=?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	if _, err := stmt.Exec(p.Name, p.Status, p.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (s *Server) getPetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("petId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var p models.Pet
	stmt, err := s.db.Prepare(`SELECT id, name, status FROM pets WHERE id=?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	row := stmt.QueryRow(id)
	if err := row.Scan(&p.ID, &p.Name, &p.Status); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, p)
}

func (s *Server) deletePet(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("petId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	stmt, err := s.db.Prepare(`DELETE FROM pets WHERE id=?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	if _, err := stmt.Exec(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) findPetsByStatus(c *gin.Context) {
	statuses := strings.Split(c.DefaultQuery("status", "available"), ",")
	placeholders := strings.Repeat("?,", len(statuses))
	placeholders = strings.TrimRight(placeholders, ",")
	args := make([]interface{}, len(statuses))
	for i, v := range statuses {
		args[i] = strings.TrimSpace(v)
	}
	query := fmt.Sprintf(`SELECT id, name, status FROM pets WHERE status IN (%s)`, placeholders)
	stmt, err := s.db.Prepare(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var pets []models.Pet
	for rows.Next() {
		var p models.Pet
		if err := rows.Scan(&p.ID, &p.Name, &p.Status); err == nil {
			pets = append(pets, p)
		}
	}
	c.JSON(http.StatusOK, pets)
}

func (s *Server) findPetsByTags(c *gin.Context) {
	tags := strings.Split(c.Query("tags"), ",")
	if len(tags) == 0 || tags[0] == "" {
		c.JSON(http.StatusOK, []models.Pet{})
		return
	}
	placeholders := strings.Repeat("?,", len(tags))
	placeholders = strings.TrimRight(placeholders, ",")
	args := make([]interface{}, len(tags))
	for i, v := range tags {
		args[i] = strings.TrimSpace(v)
	}
	query := fmt.Sprintf(`SELECT DISTINCT p.id, p.name, p.status FROM pets p JOIN pet_tags pt ON p.id=pt.pet_id JOIN tags t ON pt.tag_id=t.id WHERE t.name IN (%s)`, placeholders)
	stmt, err := s.db.Prepare(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var pets []models.Pet
	for rows.Next() {
		var p models.Pet
		if err := rows.Scan(&p.ID, &p.Name, &p.Status); err == nil {
			pets = append(pets, p)
		}
	}
	c.JSON(http.StatusOK, pets)
}

func (s *Server) updatePetWithForm(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("petId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	name := c.Query("name")
	status := c.Query("status")
	if name != "" {
		stmt, err := s.db.Prepare(`UPDATE pets SET name=? WHERE id=?`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer stmt.Close()
		if _, err := stmt.Exec(name, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if status != "" {
		stmt, err := s.db.Prepare(`UPDATE pets SET status=? WHERE id=?`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer stmt.Close()
		if _, err := stmt.Exec(status, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	s.getPetByID(c)
}

func (s *Server) uploadFile(c *gin.Context) {
	_, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"filename": header.Filename})
}

func (s *Server) getInventory(c *gin.Context) {
	stmt, err := s.db.Prepare(`SELECT status, COUNT(*) FROM pets GROUP BY status`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	inv := map[string]int{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err == nil {
			inv[status] = count
		}
	}
	c.JSON(http.StatusOK, inv)
}

func (s *Server) placeOrder(c *gin.Context) {
	var o models.Order
	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stmt, err := s.db.Prepare(`INSERT INTO orders (pet_id, quantity, ship_date, status, complete) VALUES (?,?,?,?,?)`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	res, err := stmt.Exec(o.PetID, o.Quantity, o.ShipDate.Format(time.RFC3339), o.Status, o.Complete)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	o.ID = id
	c.JSON(http.StatusOK, o)
}

func (s *Server) getOrderByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("orderId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var o models.Order
	stmt, err := s.db.Prepare(`SELECT id, pet_id, quantity, ship_date, status, complete FROM orders WHERE id=?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	row := stmt.QueryRow(id)
	var ship string
	if err := row.Scan(&o.ID, &o.PetID, &o.Quantity, &ship, &o.Status, &o.Complete); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	if ship != "" {
		o.ShipDate, _ = time.Parse(time.RFC3339, ship)
	}
	c.JSON(http.StatusOK, o)
}

func (s *Server) deleteOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("orderId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	stmt, err := s.db.Prepare(`DELETE FROM orders WHERE id=?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	if _, err := stmt.Exec(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) createUser(c *gin.Context) {
	var u models.User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stmt, err := s.db.Prepare(`INSERT INTO users (username, first_name, last_name, email, password, phone, user_status) VALUES (?,?,?,?,?,?,?)`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	res, err := stmt.Exec(u.Username, u.FirstName, u.LastName, u.Email, u.Password, u.Phone, u.UserStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	u.ID = id
	c.JSON(http.StatusOK, u)
}

func (s *Server) createUsersWithListInput(c *gin.Context) {
	var users []models.User
	if err := c.ShouldBindJSON(&users); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tx, err := s.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	stmt, err := tx.Prepare(`INSERT INTO users (username, first_name, last_name, email, password, phone, user_status) VALUES (?,?,?,?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	for i, u := range users {
		res, err := stmt.Exec(u.Username, u.FirstName, u.LastName, u.Email, u.Password, u.Phone, u.UserStatus)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id, _ := res.LastInsertId()
		users[i].ID = id
	}
	tx.Commit()
	c.JSON(http.StatusOK, users)
}

func (s *Server) loginUser(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	var id int64
	stmt, err := s.db.Prepare(`SELECT id FROM users WHERE username=? AND password=?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(username, password).Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username/password"})
		return
	}
	c.String(http.StatusOK, "logged in user %s", username)
}

func (s *Server) logoutUser(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (s *Server) getUserByName(c *gin.Context) {
	username := c.Param("username")
	var u models.User
	stmt, err := s.db.Prepare(`SELECT id, username, first_name, last_name, email, password, phone, user_status FROM users WHERE username=?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	row := stmt.QueryRow(username)
	if err := row.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.Phone, &u.UserStatus); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, u)
}

func (s *Server) updateUser(c *gin.Context) {
	username := c.Param("username")
	var u models.User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stmt, err := s.db.Prepare(`UPDATE users SET username=?, first_name=?, last_name=?, email=?, password=?, phone=?, user_status=? WHERE username=?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	if _, err := stmt.Exec(u.Username, u.FirstName, u.LastName, u.Email, u.Password, u.Phone, u.UserStatus, username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.getUserByName(c)
}

func (s *Server) deleteUser(c *gin.Context) {
	username := c.Param("username")
	stmt, err := s.db.Prepare(`DELETE FROM users WHERE username=?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()
	if _, err := stmt.Exec(username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
