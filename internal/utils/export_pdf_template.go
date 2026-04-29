package utils

import (
	"bytes"
	"html/template"
)

// PDFTemplateData — data untuk render HTML template PDF
type PDFTemplateData struct {
	Title      string
	Date       string
	Headers    []string
	Rows       [][]string
	TotalData  int
}

const pdfTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: 'Segoe UI', system-ui, sans-serif;
    color: #1a1a2e; background: #fff; padding: 24px;
  }
  .header {
    display: flex; justify-content: space-between; align-items: center;
    border-bottom: 3px solid #daa520; padding-bottom: 16px; margin-bottom: 20px;
  }
  .header h1 { font-size: 18px; color: #1a1a2e; }
  .header .meta { font-size: 11px; color: #666; text-align: right; }
  table {
    width: 100%; border-collapse: collapse; font-size: 10px;
  }
  th {
    background: #1a1a2e; color: #daa520; padding: 8px 10px;
    text-align: left; font-weight: 600; text-transform: uppercase;
    letter-spacing: 0.5px; font-size: 9px;
  }
  td {
    padding: 6px 10px; border-bottom: 1px solid #e5e7eb;
  }
  tr:nth-child(even) td { background: #f9fafb; }
  .footer {
    margin-top: 16px; padding-top: 12px; border-top: 1px solid #e5e7eb;
    font-size: 10px; color: #999; text-align: right;
  }
</style>
</head>
<body>
  <div class="header">
    <h1>{{.Title}}</h1>
    <div class="meta">
      <div>Tanggal: {{.Date}}</div>
      <div>Total: {{.TotalData}} data</div>
    </div>
  </div>
  <table>
    <thead>
      <tr>{{range .Headers}}<th>{{.}}</th>{{end}}</tr>
    </thead>
    <tbody>
      {{range .Rows}}<tr>{{range .}}<td>{{.}}</td>{{end}}</tr>{{end}}
    </tbody>
  </table>
  <div class="footer">Dicetak dari HRIS Wafa — {{.Date}}</div>
</body>
</html>`

// RenderPDFHTML — render template ke HTML string
func RenderPDFHTML(data PDFTemplateData) (string, error) {
	t, err := template.New("pdf").Parse(pdfTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
