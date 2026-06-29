package handler

import (
	"fmt"
	"net/http"
	"time"

	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

type ReportHandler struct{}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

// GET /api/v1/branches/{id}/reports/sales?start=&end=
func (h *ReportHandler) Sales(w http.ResponseWriter, r *http.Request) {
	branchID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	rows, totalSales, totalDiscount, totalNet, totalTx, err := repository.GetSalesReport(branchID, start, end)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []model.SalesReportRow{}
	}

	resp := model.SalesReportResponse{}
	resp.Period.Start = start.Format(time.RFC3339)
	resp.Period.End = end.Format(time.RFC3339)
	resp.Rows = rows
	resp.TotalSales = totalSales
	resp.TotalDiscount = totalDiscount
	resp.TotalNet = totalNet
	resp.TotalTransactions = totalTx

	writeJSON(w, http.StatusOK, resp)
}

// GET /api/v1/branches/{id}/reports/stock
func (h *ReportHandler) Stock(w http.ResponseWriter, r *http.Request) {
	branchID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	rows, err := repository.GetStockReport(branchID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []model.StockReportRow{}
	}

	resp := model.StockReportResponse{
		Rows:       rows,
		TotalItems: len(rows),
	}

	writeJSON(w, http.StatusOK, resp)
}

// GET /api/v1/branches/{id}/reports/profit-loss?start=&end=
func (h *ReportHandler) ProfitLoss(w http.ResponseWriter, r *http.Request) {
	branchID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	rows, summary, err := repository.GetProfitLossReport(branchID, start, end)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []model.ProfitLossRow{}
	}

	resp := model.ProfitLossReportResponse{}
	resp.Period.Start = start.Format(time.RFC3339)
	resp.Period.End = end.Format(time.RFC3339)
	resp.Rows = rows
	resp.Summary = summary

	writeJSON(w, http.StatusOK, resp)
}

// GET /api/v1/branches/{id}/reports/sales.pdf?start=&end=
func (h *ReportHandler) SalesPDF(w http.ResponseWriter, r *http.Request) {
	branchID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	data, err := repository.GetSalesPDFData(branchID, start, end)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// ── Generate PDF with gofpdf ──
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Title
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 12, fmt.Sprintf("Laporan Penjualan (%s - %s)",
		start.Format("02 Jan 2006"), end.Format("02 Jan 2006")), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// ── Table header ──
	colWidths := []float64{30, 55, 18, 25, 30, 25, 25}
	headers := []string{"Tanggal", "Produk", "Qty", "Harga", "Subtotal", "Pajak", "Total"}

	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(68, 114, 196)
	pdf.SetTextColor(255, 255, 255)
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// ── Table body ──
	pdf.SetTextColor(30, 30, 30)
	pdf.SetFont("Helvetica", "", 8)

	var grandTotal float64
	for _, row := range data {
		pdf.CellFormat(colWidths[0], 7, row.Date[:10], "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[1], 7, row.ProductName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[2], 7, fmt.Sprintf("%d", row.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[3], 7, fmt.Sprintf("%.0f", row.Price), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[4], 7, fmt.Sprintf("%.0f", row.Subtotal), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[5], 7, fmt.Sprintf("%.0f", row.TaxAmount), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[6], 7, fmt.Sprintf("%.0f", row.Total), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
		grandTotal = row.Total
	}

	// ── Grand total row ──
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(220, 220, 220)
	// Merge cells: colspan 6, then total
	for i := 0; i < 6; i++ {
		pdf.CellFormat(colWidths[i], 8, "", "1", 0, "", true, 0, "")
	}
	pdf.CellFormat(colWidths[6], 8, fmt.Sprintf("%.0f", grandTotal), "1", 1, "R", true, 0, "")

	// ── Write response ──
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="sales_report_%s.pdf"`, start.Format("2006-01-02")))

	if err := pdf.Output(w); err != nil {
		// Headers already sent; can only log
		_ = err
	}
}

// ─── Helper ───

func parseTimeRange(r *http.Request) (time.Time, time.Time, error) {
	q := r.URL.Query()
	startStr := q.Get("start")
	endStr := q.Get("end")

	now := time.Now()

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			start, err = time.Parse("2006-01-02", startStr)
			if err != nil {
				return start, end, err
			}
		}
	} else {
		start = now.Truncate(24 * time.Hour).AddDate(0, 0, -7) // default: last 7 days
	}

	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			end, err = time.Parse("2006-01-02", endStr)
			if err != nil {
				return start, end, err
			}
		}
	} else {
		end = now
	}

	return start, end, nil
}
