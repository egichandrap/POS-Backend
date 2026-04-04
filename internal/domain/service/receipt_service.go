package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
)

// ReceiptService generates receipts for transactions
type ReceiptService struct {
	storeName    string
	storeAddress string
	storePhone   string
	storeCode    string
}

// NewReceiptService creates a new ReceiptService
func NewReceiptService(storeName, storeAddress, storePhone, storeCode string) *ReceiptService {
	return &ReceiptService{
		storeName:    storeName,
		storeAddress: storeAddress,
		storePhone:   storePhone,
		storeCode:    storeCode,
	}
}

// GenerateTextReceipt generates a text-based receipt
func (s *ReceiptService) GenerateTextReceipt(transaction *model.Transaction) string {
	var builder strings.Builder

	// Header
	builder.WriteString(s.formatReceiptHeader())
	builder.WriteString("\n")

	// Store info
	builder.WriteString(s.formatStoreInfo())
	builder.WriteString("\n")

	// Transaction info
	builder.WriteString(s.formatTransactionInfo(transaction))
	builder.WriteString("\n")

	// Items
	builder.WriteString(s.formatItems(transaction.Items()))
	builder.WriteString("\n")

	// Totals
	builder.WriteString(s.formatTotals(transaction))
	builder.WriteString("\n")

	// Payment info
	builder.WriteString(s.formatPaymentInfo(transaction))
	builder.WriteString("\n")

	// Footer
	builder.WriteString(s.formatFooter())

	return builder.String()
}

func (s *ReceiptService) formatReceiptHeader() string {
	width := 40
	line := strings.Repeat("=", width)
	return fmt.Sprintf("%s\n%s\n%s", line, s.centerText(s.storeName, width), line)
}

func (s *ReceiptService) formatStoreInfo() string {
	return fmt.Sprintf("%s\nTelp: %s\nKode: %s",
		s.storeAddress,
		s.storePhone,
		s.storeCode,
	)
}

func (s *ReceiptService) formatTransactionInfo(transaction *model.Transaction) string {
	line := strings.Repeat("-", 40)
	return fmt.Sprintf("%s\nNo: %s\nTanggal: %s\nKasir: %s",
		line,
		transaction.TransactionNo(),
		transaction.CreatedAt().Format("02 Jan 2006 15:04:05"),
		transaction.CashierName(),
	)
}

func (s *ReceiptService) formatItems(items []model.TransactionItem) string {
	var builder strings.Builder
	line := strings.Repeat("-", 40)
	builder.WriteString(line)
	builder.WriteString("\n")
	
	for _, item := range items {
		builder.WriteString(fmt.Sprintf("%-20s %10s\n", item.ProductName(), fmt.Sprintf("%.0f", item.UnitPrice())))
		builder.WriteString(fmt.Sprintf("  %d x %.0f %15s\n", 
			item.Quantity(), 
			item.UnitPrice(),
			fmt.Sprintf("%.0f", item.Subtotal())))
	}
	
	builder.WriteString(line)
	return builder.String()
}

func (s *ReceiptService) formatTotals(transaction *model.Transaction) string {
	var builder strings.Builder
	
	builder.WriteString(fmt.Sprintf("Subtotal: %18.0f\n", transaction.Subtotal()))
	
	if transaction.DiscountAmount() > 0 {
		builder.WriteString(fmt.Sprintf("Diskon (%.0f%%): %15.0f\n", 
			transaction.DiscountPercent(), 
			transaction.DiscountAmount()))
	}
	
	if transaction.TaxAmount() > 0 {
		builder.WriteString(fmt.Sprintf("Pajak (%.0f%%): %16.0f\n", 
			transaction.TaxPercent(), 
			transaction.TaxAmount()))
	}
	
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString(fmt.Sprintf("\nTOTAL: %21.0f\n", transaction.TotalAmount()))
	
	return builder.String()
}

func (s *ReceiptService) formatPaymentInfo(transaction *model.Transaction) string {
	line := strings.Repeat("-", 40)
	return fmt.Sprintf("%s\nMetode: %s\nBayar: %18.0f\nKembali: %18.0f",
		line,
		transaction.PaymentMethod(),
		transaction.PaymentAmount(),
		transaction.ChangeAmount(),
	)
}

func (s *ReceiptService) formatFooter() string {
	line := strings.Repeat("=", 40)
	return fmt.Sprintf("%s\nTerima kasih atas kunjungan Anda!\n%s\nBarang yang sudah dibeli tidak dapat dikembalikan",
		line,
		line,
	)
}

func (s *ReceiptService) centerText(text string, width int) string {
	padding := (width - len(text)) / 2
	if padding < 0 {
		padding = 0
	}
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", padding)
}

// GenerateHTMLReceipt generates an HTML-based receipt for printing
func (s *ReceiptService) GenerateHTMLReceipt(transaction *model.Transaction) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Receipt - %s</title>
    <style>
        body { font-family: 'Courier New', monospace; max-width: 300px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; border-bottom: 2px solid #000; padding-bottom: 10px; }
        .store-info { margin: 10px 0; }
        .items { margin: 10px 0; }
        .item { margin: 5px 0; }
        .totals { border-top: 2px solid #000; padding-top: 10px; margin-top: 10px; }
        .total { font-size: 1.2em; font-weight: bold; }
        .footer { text-align: center; margin-top: 20px; border-top: 2px solid #000; padding-top: 10px; }
        @media print { body { margin: 0; } }
    </style>
</head>
<body>
    <div class="header">
        <h2>%s</h2>
        <p>%s<br>%s</p>
    </div>
    
    <div class="store-info">
        <p>No: %s<br>
        Tanggal: %s<br>
        Kasir: %s</p>
    </div>
    
    <div class="items">
        %s
    </div>
    
    <div class="totals">
        <p>Subtotal: %.0f<br>
        %s%s
        <span class="total">TOTAL: %.0f</span></p>
    </div>
    
    <div>
        <p>Metode: %s<br>
        Bayar: %.0f<br>
        Kembali: %.0f</p>
    </div>
    
    <div class="footer">
        <p>Terima kasih atas kunjungan Anda!</p>
        <p>Barang yang sudah dibeli tidak dapat dikembalikan</p>
    </div>
</body>
</html>`

	// Generate items HTML
	var itemsHTML string
	for _, item := range transaction.Items() {
		itemsHTML += fmt.Sprintf(`
        <div class="item">
            <div>%s - %.0f</div>
            <div>%d x %.0f = %.0f</div>
        </div>`, item.ProductName(), item.UnitPrice(), item.Quantity(), item.UnitPrice(), item.Subtotal())
	}

	// Generate discount HTML
	discountHTML := ""
	if transaction.DiscountAmount() > 0 {
		discountHTML = fmt.Sprintf("Diskon (%.0f%%): %.0f<br>", transaction.DiscountPercent(), transaction.DiscountAmount())
	}

	// Generate tax HTML
	taxHTML := ""
	if transaction.TaxAmount() > 0 {
		taxHTML = fmt.Sprintf("Pajak (%.0f%%): %.0f<br>", transaction.TaxPercent(), transaction.TaxAmount())
	}

	return fmt.Sprintf(html,
		transaction.TransactionNo(),
		s.storeName,
		s.storeAddress,
		s.storePhone,
		transaction.TransactionNo(),
		transaction.CreatedAt().Format("02 Jan 2006 15:04:05"),
		transaction.CashierName(),
		itemsHTML,
		transaction.Subtotal(),
		discountHTML,
		taxHTML,
		transaction.TotalAmount(),
		transaction.PaymentMethod(),
		transaction.PaymentAmount(),
		transaction.ChangeAmount(),
	)
}

// PrintReceipt prints receipt to printer (placeholder)
func (s *ReceiptService) PrintReceipt(transaction *model.Transaction, printerName string) error {
	// TODO: Implement actual printer integration
	// For now, just generate text receipt
	receipt := s.GenerateTextReceipt(transaction)
	fmt.Printf("Printing to: %s\n%s\n", printerName, receipt)
	return nil
}

// SaveReceipt saves receipt to file (placeholder)
func (s *ReceiptService) SaveReceipt(transaction *model.Transaction, format string) (string, error) {
	// TODO: Implement file saving (PDF, TXT, HTML)
	filename := fmt.Sprintf("receipt_%s_%s.%s",
		transaction.TransactionNo(),
		time.Now().Format("20060102_150405"),
		format,
	)
	
	return filename, nil
}
