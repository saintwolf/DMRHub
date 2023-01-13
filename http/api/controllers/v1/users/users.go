package users

import (
	"net/http"
	"strings"

	"github.com/USA-RedDragon/dmrserver-in-a-box/http/api/apimodels"
	"github.com/USA-RedDragon/dmrserver-in-a-box/http/api/utils"
	"github.com/USA-RedDragon/dmrserver-in-a-box/models"
	"github.com/USA-RedDragon/dmrserver-in-a-box/userdb"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

func GETUsers(c *gin.Context) {
}

// Registration is JSON data from the frontend
func POSTUser(c *gin.Context) {
	db := c.MustGet("DB").(*gorm.DB)
	var json apimodels.Registration
	err := c.ShouldBindJSON(&json)
	if err != nil {
		klog.Errorf("POSTUser: JSON data is invalid: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON data is invalid"})
	} else {
		registeredDMRID := false
		matchesCallsign := false
		// Check DMR ID is in the database
		userDB := *userdb.GetDMRUsers()
		for _, user := range userDB {
			if user.ID == json.DMRId {
				registeredDMRID = true
			}

			if strings.EqualFold(user.Callsign, json.Callsign) {
				matchesCallsign = true
				if registeredDMRID {
					break
				}
			}
		}
		if !registeredDMRID || !matchesCallsign {
			c.JSON(http.StatusBadRequest, gin.H{"error": "DMR ID is not registered or Callsign does not match"})
			return
		}
		isValid, errString := json.IsValidUsername()
		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": errString})
			return
		}

		// Check that password isn't a zero string
		if json.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password cannot be blank"})
			return
		}

		// Check if the username is already taken
		var user models.User
		db.Find(&user, "username = ?", json.Username)
		if db.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": db.Error.Error()})
			return
		} else if user.ID != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username is already taken"})
			return
		}

		// argon2 the password
		hashedPassword := utils.HashPassword(json.Password)

		// store the user in the database with Active = false
		user = models.User{
			Username: json.Username,
			Password: hashedPassword,
			Callsign: json.Callsign,
			ID:       json.DMRId,
			Approved: false,
			Admin:    false,
		}
		db.Create(&user)
		if db.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": db.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User created, please wait for admin approval"})
	}
}

func POSTUserAdmins(c *gin.Context) {
	db := c.MustGet("DB").(*gorm.DB)
	var json apimodels.UserAdmins
	err := c.ShouldBindJSON(&json)
	if err != nil {
		klog.Errorf("POSTUser: JSON data is invalid: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON data is invalid"})
	} else {
		// Grab the user from the database
		var user models.User
		db.Find(&user, "id = ?", json.ID)
		if db.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": db.Error.Error()})
			return
		}
		if user.ID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
			return
		}
		user.Admin = json.Admin
		db.Save(&user)
		if db.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": db.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User promoted"})
	}
}

func POSTUserApprove(c *gin.Context) {
	db := c.MustGet("DB").(*gorm.DB)
	var json apimodels.UserApprove
	err := c.ShouldBindJSON(&json)
	if err != nil {
		klog.Errorf("POSTUserApprove: JSON data is invalid: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON data is invalid"})
	} else {
		// Grab the user from the database
		var user models.User
		db.Find(&user, "id = ?", json.ID)
		if db.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": db.Error.Error()})
			return
		}
		if user.ID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
			return
		}
		user.Approved = true
		db.Save(&user)
		if db.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": db.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User approved"})
	}
}

func GETUser(c *gin.Context) {
}

func GETUserAdmins(c *gin.Context) {
}

func PATCHUser(c *gin.Context) {
}

func DELETEUser(c *gin.Context) {
}
