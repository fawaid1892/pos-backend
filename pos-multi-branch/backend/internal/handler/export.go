package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"pos-multi-branch/backend/internal/repository"

	"github.com/google/uuid"
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
	fmt.Fprintln(w, "Date,Customer,Items,Subtotal,Discount,Total,Cash,Change")

	for _, row := range data {
		// Escape commas in item string
		items := strings.ReplaceAll(row.Items, "\"", "\"\"")
		fmt.Fprintf(w, "%s,\"%s\",\"%s\",%.2f,%.2f,%.2f,%.2f,%.2f\n",
			row.Date, row.CustomerName, items, row.Subtotal, row.Discount, row.Total, row.Cash, row.Change)
	}
}

func (h *ExportHandler) exportXLSX(w http.ResponseWriter, data []repository.SalesExportRow, start, end time.Time) {
	// Simple XLSX generation using CSV-like format wrapped in XML
	// For production, use a library like excelize
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="sales_export_%s.xlsx"`, start.Format("2006-01-02")))

	// Build a minimal XLSX XML table
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<?mso-application progid="Excel.Sheet"?>
<Workbook xmlns="urn:schemas-microsoft-com:office:spreadsheet"
 xmlns:o="urn:schemas-microsoft-com:office:office"
 xmlns:x="urn:schemas-microsoft-com:office:excel"
 xmlns:ss="urn:schemas-microsoft-com:office:spreadsheet">
 <Worksheet ss:Name="Sales">
  <Table>
   <Row>
    <Cell><Data ss:Type="String">Date</Data></Cell>
    <Cell><Data ss:Type="String">Customer</Data></Cell>
    <Cell><Data ss:Type="String">Items</Data></Cell>
    <Cell><Data ss:Type="String">Subtotal</Data></Cell>
    <Cell><Data ss:Type="String">Discount</Data></Cell>
    <Cell><Data ss:Type="String">Total</Data></Cell>
    <Cell><Data ss:Type="String">Cash</Data></Cell>
    <Cell><Data ss:Type="String">Change</Data></Cell>
   </Row>`

	for _, row := range data {
		xml += fmt.Sprintf(`
   <Row>
    <Cell><Data ss:Type="String">%s</Data></Cell>
    <Cell><Data ss:Type="String">%s</Data></Cell>
    <Cell><Data ss:Type="String">%s</Data></Cell>
    <Cell><Data ss:Type="Number">%.2f</Data></Cell>
    <Cell><Data ss:Type="Number">%.2f</Data></Cell>
    <Cell><Data ss:Type="Number">%.2f</Data></Cell>
    <Cell><Data ss:Type="Number">%.2f</Data></Cell>
    <Cell><Data ss:Type="Number">%.2f</Data></Cell>
   </Row>`,
			escapeXML(row.Date), escapeXML(row.CustomerName), escapeXML(row.Items),
			row.Subtotal, row.Discount, row.Total, row.Cash, row.Change)
	}

	xml += `
  </Table>
 </Worksheet>
</Workbook>`

	fmt.Fprint(w, xml)
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
