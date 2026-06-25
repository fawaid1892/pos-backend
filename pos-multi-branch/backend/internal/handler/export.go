package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"pos-multi-branch/backend/internal/repository"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type ExportHandler struct{}

func NewExportHandler() *ExportHandler {
	return &ExportHandler{}
}

// GET /api/v1/branches/{id}/reports/sales/export?format=pdf|xlsx
func (h *ExportHandler) SalesExport(w http.ResponseWriter, r *http.Request) {
	branchID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch id"})
		return
	}

	format := strings.ToLower(r.URL.Query().Get("format"))
	if format == "" {
		format = "csv"
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	data, err := repository.GetSalesExportData(r.Context(), branchID, start, end)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	switch format {
	case "csv":
		h.exportCSV(w, data, start, end)
	case "xlsx":
		h.exportXLSX(w, data, start, end)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported format, use csv or xlsx"})
	}
}

func (h *ExportHandler) exportCSV(w http.ResponseWriter, data []repository.SalesExportRow, start, end time.Time) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="sales_export_%s.csv"`, start.Format("2006-01-02")))

	// Write CSV header
	fmt.Fprintln(w, "No,Date,Customer,Items,Subtotal,Discount,Pajak,Total,Cash,Change")

	for i, row := range data {
		// Escape commas in item string
		items := strings.ReplaceAll(row.Items, "\"", "\"\"")
		fmt.Fprintf(w, "%d,%s,\"%s\",\"%s\",%.2f,%.2f,%.2f,%.2f,%.2f,%.2f\n",
			i+1, row.Date, row.CustomerName, items, row.Subtotal, row.Discount, row.TaxAmount, row.Total, row.Cash, row.Change)
	}
}

func (h *ExportHandler) exportXLSX(w http.ResponseWriter, data []repository.SalesExportRow, start, end time.Time) {
	f := excelize.NewFile()
	defer f.Close()

	// Rename default sheet
	f.SetSheetName("Sheet1", "Sales")

	// ─── Headers ───
	headers := []string{"No", "Tanggal", "Pelanggan", "Produk", "Subtotal", "Diskon", "Pajak", "Total", "Tunai", "Kembalian"}
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sales", cell, h)
	}
	f.SetCellStyle("Sales", "A1", "J1", styleHeader)

	// ─── Data ───
	styleNumber, _ := f.NewStyle(&excelize.Style{
		NumFmt: 4, // #,##0.00 - 2 decimal places
	})

	for i, row := range data {
		rowNum := i + 2
		f.SetCellInt("Sales", fmt.Sprintf("A%d", rowNum), i+1)
		f.SetCellValue("Sales", fmt.Sprintf("B%d", rowNum), row.Date)
		f.SetCellValue("Sales", fmt.Sprintf("C%d", rowNum), row.CustomerName)
		f.SetCellValue("Sales", fmt.Sprintf("D%d", rowNum), row.Items)
		f.SetCellFloat("Sales", fmt.Sprintf("E%d", rowNum), row.Subtotal, 2, 64)
		f.SetCellFloat("Sales", fmt.Sprintf("F%d", rowNum), row.Discount, 2, 64)
		f.SetCellFloat("Sales", fmt.Sprintf("G%d", rowNum), row.TaxAmount, 2, 64)
		f.SetCellFloat("Sales", fmt.Sprintf("H%d", rowNum), row.Total, 2, 64)
		f.SetCellFloat("Sales", fmt.Sprintf("I%d", rowNum), row.Cash, 2, 64)
		f.SetCellFloat("Sales", fmt.Sprintf("J%d", rowNum), row.Change, 2, 64)

		// Apply 2-decimal format to number columns
		for _, col := range []string{"E", "F", "G", "H", "I", "J"} {
			f.SetCellStyle("Sales", fmt.Sprintf("%s%d", col, rowNum), fmt.Sprintf("%s%d", col, rowNum), styleNumber)
		}
	}

	// ─── Auto-fit column widths ───
	colWidths := []float64{5, 22, 18, 30, 14, 12, 12, 14, 14, 14}
	for i, w := range colWidths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth("Sales", col, col, w)
	}

	// ─── Write response ───
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="sales_export_%s.xlsx"`, start.Format("2006-01-02")))

	if err := f.Write(w); err != nil {
		// If writing fails, fall back to error JSON (headers already sent though)
		_ = err
	}
}
