package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Computer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	MAC  string `json:"mac"`
}

var db *sql.DB

// Initialize the database
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./computers.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	createTable := `
	CREATE TABLE IF NOT EXISTS computers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		mac TEXT
	);
	`
	if _, err = db.Exec(createTable); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

// SendWOLPacket sends a Wake-on-LAN magic packet to the given MAC address.
func SendWOLPacket(mac string) error {
	hwAddr, err := net.ParseMAC(mac)
	if err != nil {
		return err
	}

	// Magic packet is 6x 0xFF followed by 16x the MAC address
	packet := make([]byte, 102)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	for i := 1; i <= 16; i++ {
		copy(packet[i*6:], hwAddr)
	}

	addr := net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 9,
	}

	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}

func main() {
	initDB()

	r := gin.Default()

	r.GET("/computers", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, mac FROM computers")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var computers []Computer
		for rows.Next() {
			var comp Computer
			if err := rows.Scan(&comp.ID, &comp.Name, &comp.MAC); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			computers = append(computers, comp)
		}

		c.JSON(http.StatusOK, computers)
	})

	r.POST("/computers", func(c *gin.Context) {
		var comp Computer
		if err := c.ShouldBindJSON(&comp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec("INSERT INTO computers (name, mac) VALUES (?, ?)", comp.Name, comp.MAC)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "Computer added"})
	})

	r.DELETE("/computers/:id", func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM computers WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "Computer deleted"})
	})

	r.POST("/wol/:mac", func(c *gin.Context) {
		mac := c.Param("mac")
		if err := SendWOLPacket(mac); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "WOL packet sent"})
	})

	r.Run(":8080") // Run on port 8080
}
