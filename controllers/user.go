package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"project/database"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"project/models"
)

type UserRepo struct {
	Db *gorm.DB
}

type StatusUser struct {
	Status string
}

func NewUserController() *UserRepo {
	db := database.InitDb()
	db.AutoMigrate(&models.User{})
	return &UserRepo{Db: db}
}

// create User
func (repository *UserRepo) CreateUser(c *gin.Context) {
	var User models.User
	c.BindJSON(&User)
	err := models.CreateUser(repository.Db, &User)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, User)
}

func (repository *UserRepo) Register(c *gin.Context) {
	var user models.User
	c.BindJSON(&user)
	// Hash password before storing it in the database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash the password"})
		return
	}
	user.Password = string(hashedPassword)

	user.ActivationCode = generateActivationCode()
	user.Active = true // Set the Active field to 1 (true)
	err = models.Register(repository.Db, &user)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	/*
		if err := repository.Db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	*/

	if err := sendActivationEmail(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send activation email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func generateActivationCode() string {
	token := make([]byte, 32)
	rand.Read(token)
	return hex.EncodeToString(token)
}

func sendActivationEmail(user models.User) error {
	// Kullanıcının e-posta adresini alın
	recipientEmail := user.Email

	// E-posta konusu ve içeriği oluşturun
	subject := "Hesap Aktivasyonu Tamamlandı"
	body := "Merhaba " + user.Username + ",\n\nHesabınız başarıyla aktive edildi. Artık giriş yapabilir ve sitemizi kullanabilirsiniz.\n\nTeşekkürler,\nSitemiz Ekibi"

	// Get Sender Name and Sender Email Address from environment variables
	senderName := os.Getenv("SENDER_NAME")
	senderEmailVisible := os.Getenv("SENDER_EMAIL_VISIBLE")

	// E-postanın gönderici adresi ve bilgileri
	senderEmail := os.Getenv("SENDER_EMAIL")
	senderPassword := os.Getenv("SENDER_PASSWORD")
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		log.Printf("SMTP_PORT is not a valid integer: %s\n", err)
		return err
	}

	// E-posta gövdesini ayarlayın
	//message := []byte("From: " + senderName + "<" + senderEmailVisible + ">" + "\r\n" + "To: " + recipientEmail + "\r\n" + "Subject: " + subject + "\r\n" + "\r\n" + body + "\r\n")
	message := []byte("To: " + recipientEmail + "\r\n" + "From: \"" + senderName + "\" <" + senderEmailVisible + ">\r\n" + "Subject: " + subject + "\r\n" + "\r\n" + body + "\r\n")

	// SMTP sunucusuna bağlanın ve e-postayı gönderin
	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpServer)
	err = smtp.SendMail(smtpServer+":"+strconv.Itoa(smtpPort), auth, senderEmail, []string{recipientEmail}, message)
	if err != nil {
		log.Printf("Error sending activation email to %s: %s\n", recipientEmail, err)
		return err
	}

	return nil
}

func AuthMiddleware(tokenRepo *TokenRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token not provided"})
			return
		}

		// Strip "Bearer " prefix
		if len(tokenString) > 7 && strings.ToUpper(tokenString[0:7]) == "BEARER " {
			tokenString = tokenString[7:]
		}

		fmt.Println("Token string:", tokenString) // logging added here

		// Get the token from the database based on the token string
		tokenObj, err := tokenRepo.GetTokenByTokenString(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Token not validated"})
			return
		}

		// Check if the token is valid based on its start and expiry dates
		if time.Now().Before(*tokenObj.StartingDate) || time.Now().After(*tokenObj.EndingDate) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			return
		}

		// Set user ID in context for further use
		c.Set("user_id", tokenObj.UserID)
		c.Next()

	}
}

/*
	func (repository *UserRepo) GetUsernameFromContext(c *gin.Context) (string, error) {
		// Get the user ID from the context
		userID, exists := c.Get("user_id")
		if !exists {
			return "", errors.New("User ID not found in context")
		}

		// Get the username from the user ID
		var user models.User
		err := models.GetUser(repository.Db, &user, userID.(string))
		if err != nil {
			return "", err
		}

		// Return the username
		c.JSON(http.StatusOK, gin.H{"username": user.Username})
		return user.Username, nil

}
*/
func (repository *UserRepo) Login(c *gin.Context) {
	var user models.User
	c.BindJSON(&user)
	email := user.Email
	password := user.Password
	err := models.Login(repository.Db, &user, email)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	/*
		// Retrieve user record from database
		var dbUser models.User
		if err := repository.Db.Where("username = ?", user.Username).First(&dbUser).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
	*/

	// Compare hashed passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.Active {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is not activated. Please activate your account."})
		return
	}

	token, err := models.CreateToken(repository.Db, user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	/*
		// Set session cookie
		session := sessions.Default(c)
		session.Set("user_id", user.ID)
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	*/

	c.JSON(http.StatusOK, gin.H{"token": token.Token, "start": token.StartingDate, "expiry": token.EndingDate, "username": user.Username})
	//c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully"})
}

// Book a ticket
func (repository *UserRepo) BookTicket(c *gin.Context) {
	// Get user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
		return
	}

	// Get the ticket ID from the request parameters
	ticketID := c.Param("ticket_id")

	// Retrieve the ticket from the database
	var ticket models.Ticket
	err := models.GetTicket(repository.Db, &ticket, ticketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	// Check if the ticket is available
	if ticket.NofSeats == "0" {
		c.JSON(http.StatusConflict, gin.H{"error": "Ticket is not available"})
		return
	}

	// Convert the number of seats to an integer
	numSeats, err := strconv.Atoi(ticket.NofSeats)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	// Decrement the number of seats available in the ticket
	numSeats -= 1
	ticket.NofSeats = strconv.Itoa(numSeats)

	// Update the ticket in the database
	err = models.UpdateTicket(repository.Db, &ticket, ticketID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var user models.User
	getUsererr := models.GetUser(repository.Db, &user, strconv.Itoa(userID.(int)))
	if getUsererr != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": getUsererr})
		return
	}

	// Create a new booked ticket record
	bookedTicket := models.BTicket{
		TicketID: ticket.ID,
		UserID:   userID.(int),
	}

	// Create the booked ticket in the database
	err = models.CreateBTicket(repository.Db, &bookedTicket)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	// Send ticket details email
	if err := sendTicketDetailsEmail(user, ticket, bookedTicket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send ticket details email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket booked successfully"})
}

func sendTicketDetailsEmail(user models.User, ticket models.Ticket, bookedTicket models.BTicket) error {
	// Retrieve the necessary information
	recipientEmail := user.Email
	ticketDetails := fmt.Sprintf("Bilet ID: %d\nNereden: %s\nNereye: %s\nGidiş Tarihi: %s\nDönüş Tarihi: %s\nKalkış Saati: %s\nGeri Dönüş Saati: %s\n", bookedTicket.ID, ticket.From, ticket.To, ticket.DepartureDate, ticket.ReturnDate, ticket.DHour, ticket.RHour)

	// Create the email subject and body
	subject := "Ticket Booking Confirmation"
	body := "Merhabalar " + user.Username + ",\n\nBilet satın alma işleminiz tamamlanmıştır, teşekkür ederiz. Detayları şu şekilde görebilirsiniz:\n\n" + ticketDetails + "\nİyi uçuşlar dileriz!\n\nSaygılarımızla,\nFlightapp Ekibi"

	// Get sender information from environment variables
	senderName := os.Getenv("SENDER_NAME")
	senderEmailVisible := os.Getenv("SENDER_EMAIL_VISIBLE")
	senderEmail := os.Getenv("SENDER_EMAIL")
	senderPassword := os.Getenv("SENDER_PASSWORD")
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		log.Printf("SMTP_PORT is not a valid integer: %s\n", err)
		return err
	}

	// Compose the email message
	message := []byte("To: " + recipientEmail + "\r\n" +
		"From: \"" + senderName + "\" <" + senderEmailVisible + ">\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")

	// Connect to the SMTP server and send the email
	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpServer)
	err = smtp.SendMail(smtpServer+":"+strconv.Itoa(smtpPort), auth, senderEmail, []string{recipientEmail}, message)
	if err != nil {
		log.Printf("Error sending ticket details email to %s: %s\n", recipientEmail, err)
		return err
	}

	return nil
}

// get Users
func (repository *UserRepo) GetUsers(c *gin.Context) {
	var User []models.User
	err := models.GetUsers(repository.Db, &User)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, User)
}

// get User by id
func (repository *UserRepo) GetUser(c *gin.Context) {
	id, _ := c.Params.Get("id")
	var User models.User
	err := models.GetUser(repository.Db, &User, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, User)
}

// update User
func (repository *UserRepo) UpdateUser(c *gin.Context) {
	id, _ := c.Params.Get("id")
	var User models.User
	c.BindJSON(&User)
	err := models.UpdateUser(repository.Db, &User, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, User)
}

// delete Country
func (repository *UserRepo) DeleteUser(c *gin.Context) {
	id, _ := c.Params.Get("id")
	var User models.User
	err := models.DeleteUser(repository.Db, &User, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "User deleted"})
}

func (repository *UserRepo) Logout(c *gin.Context) {
	// Get the token string from the request headers or session
	/*
	   tokenString := c.GetHeader("Authorization")
	   // Alternatively, retrieve the token from the session if you are using session-based authentication

	   // Use the TokenRepo to retrieve the token object
	   tokenObj, err := TokenRepo.GetTokenByTokenString(tokenString)
	   if err != nil {
	       c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Token not found"})
	       return
	   }

	   // Update the token's expiration date to a past date
	   tokenObj.EndingDate = time.Now().Add(-time.Hour) // Set the expiration date to 1 hour ago or any desired past date

	   // Save the updated token object in the database
	   err = repository.Db.Save(tokenObj).Error
	   if err != nil {
	       c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token"})
	       return
	   }

	   // Clear any session data or cookies associated with the user's authentication session, if applicable

	   c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
	*/
}
