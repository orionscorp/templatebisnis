package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "net/smtp"
    "os"
    "time"

    "github.com/joho/godotenv"
    "github.com/veritrans/go-midtrans"
    _ "github.com/go-sql-driver/mysql"
)

type PaymentRequest struct {
    Name          string `json:"name"`
    Phone         string `json:"phone"`
    Email         string `json:"email"`
    Bonus         bool   `json:"bonus"`
    PaymentMethod string `json:"paymentMethod"`
    TotalPrice    int    `json:"totalPrice"`
}

type MidtransNotification struct {
    OrderID           string `json:"order_id"`
    TransactionStatus string `json:"transaction_status"`
    FraudStatus       string `json:"fraud_status"`
}

const (
    PaymentTypeCreditCard      midtrans.PaymentType = "credit_card"
    PaymentTypeGopay           midtrans.PaymentType = "gopay"
    PaymentTypeBCAVA           midtrans.PaymentType = "bca_va"
    PaymentTypeBNIVA           midtrans.PaymentType = "bni_va"
    PaymentTypeEChannel        midtrans.PaymentType = "echannel"
    PaymentTypeCStoreIndomaret midtrans.PaymentType = "cstore"
    PaymentTypeCStoreAlfamart  midtrans.PaymentType = "cstore"
    PaymentTypeQris            midtrans.PaymentType = "qris"
)

var db *sql.DB

func sendEmail(subject, body, recipient string) error {
    from := os.Getenv("EMAIL_USERNAME")
    password := os.Getenv("EMAIL_PASSWORD")
    smtpHost := "smtp.gmail.com"
    smtpPort := "587"

    to := []string{recipient}

    message := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))

    auth := smtp.PlainAuth("", from, password, smtpHost)
    err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
    if err != nil {
        return err
    }
    return nil
}

func sendThankYouEmail(recipient, name string) error {
    subject := "Thanks for Paying"
    body := fmt.Sprintf("Hi %s,\n\nThank you for your payment. Your order is being processed.\n\nBest regards,\nYour Company", name)
    return sendEmail(subject, body, recipient)
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    w.Header().Set("Content-Type", "application/json")

    var req PaymentRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Initialize Snap client
    midclient := midtrans.NewClient()
    midclient.ServerKey = os.Getenv("MIDTRANS_SERVER_KEY")
    midclient.ClientKey = os.Getenv("MIDTRANS_CLIENT_KEY")
    midclient.APIEnvType = midtrans.Sandbox

    snapGateway := midtrans.SnapGateway{Client: midclient}

    // Map payment method string to midtrans.PaymentType
    paymentType := map[string]midtrans.PaymentType{
        "credit_card":      PaymentTypeCreditCard,
        "gopay":            PaymentTypeGopay,
        "bca_va":           PaymentTypeBCAVA,
        "bni_va":           PaymentTypeBNIVA,
        "echannel":         PaymentTypeEChannel,
        "cstore_indomaret": PaymentTypeCStoreIndomaret,
        "cstore_alfamart":  PaymentTypeCStoreAlfamart,
        "qris":             PaymentTypeGopay,
    }[req.PaymentMethod]

    reqSnap := &midtrans.SnapReq{
        TransactionDetails: midtrans.TransactionDetails{
            OrderID:  fmt.Sprintf("order-%d", time.Now().Unix()),
            GrossAmt: int64(req.TotalPrice),
        },
        CustomerDetail: &midtrans.CustDetail{
            FName: req.Name,
            Email: req.Email,
            Phone: req.Phone,
        },
        EnabledPayments: []midtrans.PaymentType{paymentType},
    }

    snapTokenResp, err := snapGateway.GetToken(reqSnap)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var paymentDetails string

    // Customize payment details based on payment method
    switch req.PaymentMethod {
    case "bank_transfer":
        paymentDetails = fmt.Sprintf("Please transfer to the following bank account number: %s", snapTokenResp.RedirectURL)
    case "credit_card":
        paymentDetails = fmt.Sprintf("Your credit card payment has been initiated. Please complete the payment through the following link: %s", snapTokenResp.RedirectURL)
    case "gopay":
        paymentDetails = fmt.Sprintf("Please use GoPay to complete your payment using the following link: %s", snapTokenResp.RedirectURL)
    case "bca_va":
        paymentDetails = fmt.Sprintf("Please transfer to the following BCA Virtual Account number: %s", snapTokenResp.RedirectURL)
    case "bni_va":
        paymentDetails = fmt.Sprintf("Please transfer to the following BNI Virtual Account number: %s", snapTokenResp.RedirectURL)
    case "echannel":
        paymentDetails = fmt.Sprintf("Please complete your payment using Mandiri Bill (eChannel) with the following link: %s", snapTokenResp.RedirectURL)
    case "cstore_indomaret":
        paymentDetails = fmt.Sprintf("Please complete your payment at Indomaret using the following instructions: %s", snapTokenResp.RedirectURL)
    case "cstore_alfamart":
        paymentDetails = fmt.Sprintf("Please complete your payment at Alfamart using the following instructions: %s", snapTokenResp.RedirectURL)
    case "qris":
        paymentDetails = fmt.Sprintf("Please scan the following QR code to complete your payment: %s", snapTokenResp.RedirectURL)
    default:
        paymentDetails = "Please follow the instructions to complete your payment."
    }

    // Compose the email content
    emailContent := fmt.Sprintf("Hi %s,\n\nTerima kasih atas pesanan Anda. Berikut adalah detail pesanan Anda:\n\nNama: %s\nNo. WhatsApp: %s\nEmail: %s\nBonus: %t\nTotal Harga: Rp %d\n\n%s\n\nBest regards,\nYour Company", req.Name, req.Name, req.Phone, req.Email, req.Bonus, req.TotalPrice, paymentDetails)

    err = sendEmail("Detail Pembayaran Anda", emailContent, req.Email)
    if err != nil {
        log.Printf("Failed to send email: %v", err)
        http.Error(w, "Failed to send email", http.StatusInternalServerError)
        return
    }

    // Save the transaction to the database
    _, err = db.Exec("INSERT INTO transactions (order_id, name, phone, email, total_price, payment_method, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
        reqSnap.TransactionDetails.OrderID, req.Name, req.Phone, req.Email, req.TotalPrice, req.PaymentMethod, "pending")
    if err != nil {
        log.Printf("Failed to insert transaction: %v", err)
        http.Error(w, "Failed to process transaction", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func notificationHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    w.Header().Set("Content-Type", "application/json")

    var notification MidtransNotification
    err := json.NewDecoder(r.Body).Decode(&notification)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Handle only successful transactions
    if notification.TransactionStatus == "settlement" {
        // Update transaction status in the database
        _, err := db.Exec("UPDATE transactions SET status = ? WHERE order_id = ?", "success", notification.OrderID)
        if err != nil {
            log.Printf("Failed to update transaction status: %v", err)
            http.Error(w, "Failed to update transaction status", http.StatusInternalServerError)
            return
        }

        // Retrieve the user's email and name from the database
        var email, name string
        err = db.QueryRow("SELECT email, name FROM transactions WHERE order_id = ?", notification.OrderID).Scan(&email, &name)
        if err != nil {
            log.Printf("Failed to retrieve user details: %v", err)
            http.Error(w, "Failed to retrieve user details", http.StatusInternalServerError)
            return
        }

        // Send thank you email
        err = sendThankYouEmail(email, name)
        if err != nil {
            log.Printf("Failed to send thank you email: %v", err)
            http.Error(w, "Failed to send thank you email", http.StatusInternalServerError)
            return
        }
    }

    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func main() {
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Open a database connection
    db, err = sql.Open("mysql", os.Getenv("DB_CONN_STR"))
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    // Test the database connection
    err = db.Ping()
    if err != nil {
        log.Fatalf("Failed to ping database: %v", err)
    }

    http.HandleFunc("/api/payment", paymentHandler)
    http.HandleFunc("/api/notification", notificationHandler)

    log.Println("Server is running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
