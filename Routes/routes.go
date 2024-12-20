package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	controllers "github.com/kevinmranda/AWE-Backend/Controllers"
)

// all web routes defined here
func Routes() {
	r := gin.Default()

	// CORS Middleware configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200", "http://127.0.0.1"}, // Allow frontend's origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// User Routes
	r.POST("api/createAccount", controllers.CreateAccount)
	r.POST("api/login", controllers.Login)
	r.GET("api/getUser/:id", controllers.GetUser)
	r.GET("api/getUsers/", controllers.GetUsers)
	r.POST("api/sendResetPasswordEmail/", controllers.SendResetPasswordEmail)
	r.POST("api/reset-password/:token", controllers.ResetPassword)

	// product Routes
	r.GET("api/getProduct/:filename", controllers.GetProduct)
	r.GET("api/getProducts/:id", controllers.GetProducts)
	r.GET("api/getAllProducts/", controllers.GetAllProducts)

	// Order Routes
	r.POST("api/addOrder/", controllers.AddOrder)
	r.GET("api/getOrder/:id", controllers.GetOrder)
	r.GET("api/getOrders/:id", controllers.GetOrders)

	// Payment Routes
	r.POST("api/payOrder", controllers.AddPayment)
	r.GET("api/getPayment/:id", controllers.GetPayment)
	r.GET("api/getPayments/:id", controllers.GetPayments)

	//Customer Routes
	r.POST("api/customerLogin/", controllers.CustomerAuthentication)
	r.POST("api/customerJoin/", controllers.AddCustomer)

	// User Routes
	r.DELETE("api/deleteUser/:id", controllers.DeleteUser)
	r.PUT("api/updateUser/:id", controllers.UpdateUser)
	r.PUT("api/updateUserPassword/:id", controllers.UpdateUserPassword)
	r.PUT("api/updateUserPreferences/:id", controllers.UpdateUserPreferences)
	r.GET("api/userPreferences/:id", controllers.GetUserPreferences)

	// product Routes
	r.POST("api/upload/", controllers.Upload)
	r.POST("api/insertProduct/:id", controllers.AddProduct)
	r.DELETE("api/deleteProduct/:id", controllers.DeleteProduct)
	r.PUT("api/updateProduct/:id", controllers.UpdateProduct)

	// Order Routes
	r.DELETE("api/removeOrder/:id", controllers.RemoveOrder)
	r.PUT("api/updateOrder/:id", controllers.UpdateOrder)

	// Payment Routes
	r.DELETE("api/deletePayment/:id", controllers.DeletePayment)
	r.PUT("api/updatePayment/:id", controllers.UpdatePayment)

	//AzamPay Callback
	r.POST("api/callback", controllers.AzamPayCallbackHandler)

	// //Logs Routes
	// r.GET("api/logs", controllers.GetLogs)

	r.Run()
}
