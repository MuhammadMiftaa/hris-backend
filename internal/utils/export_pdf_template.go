package utils

import (
	"bytes"
	"html/template"
)

// PDFTemplateData — data untuk render HTML template PDF
type PDFTemplateData struct {
	Title     string
	Date      string
	Headers   []string
	Rows      [][]string
	TotalData int
}

const pdfTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  @import url('https://fonts.googleapis.com/css2?family=Manrope:wght@400;500;600;700;800&display=swap');

  * { margin: 0; padding: 0; box-sizing: border-box; }

  body {
    font-family: 'Manrope', 'Segoe UI', system-ui, sans-serif;
    color: #1a0010;
    background: #fdf8fb;
    padding: 32px;
    -webkit-font-smoothing: antialiased;
  }

  /* ── Header ── */
  .header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    padding-bottom: 20px;
    margin-bottom: 24px;
    border-bottom: 2px solid rgba(209, 0, 113, 0.15);
    position: relative;
  }

  .header::after {
    content: '';
    position: absolute;
    bottom: -2px;
    left: 0;
    width: 120px;
    height: 2px;
    background: linear-gradient(90deg, #9d167c, #d10071, #dd0d89);
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .brand-logo {
    width: 42px;
    height: 42px;
    background: linear-gradient(135deg, #9d167c 0%, #d10071 50%, #dd0d89 100%);
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-weight: 800;
    font-size: 20px;
    letter-spacing: -1px;
    position: relative;
  }

  .brand-logo::before {
    content: '';
    position: absolute;
    top: 6px;
    left: 50%;
    transform: translateX(-50%);
    width: 0;
    height: 0;
    border-left: 5px solid transparent;
    border-right: 5px solid transparent;
    border-bottom: 6px solid #ffd313;
  }

  .brand-text h1 {
    font-size: 18px;
    font-weight: 800;
    color: #1a0010;
    letter-spacing: -0.3px;
    line-height: 1.2;
  }

  .brand-text .tagline {
    font-size: 10px;
    color: #7a4068;
    font-weight: 500;
    margin-top: 2px;
    letter-spacing: 0.3px;
  }

  .meta {
    text-align: right;
    font-size: 11px;
    color: #7a4068;
    line-height: 1.6;
  }

  .meta .meta-label {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 1px;
    color: #d10071;
    font-weight: 700;
    margin-bottom: 4px;
  }

  .meta .meta-value {
    font-weight: 600;
    color: #1a0010;
  }

  /* ── Title Section ── */
  .doc-title {
    margin-bottom: 20px;
  }

  .doc-title h2 {
    font-size: 22px;
    font-weight: 800;
    color: #1a0010;
    letter-spacing: -0.5px;
    line-height: 1.3;
    background: linear-gradient(135deg, #9d167c 0%, #d10071 35%, #dd0d89 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .doc-title .subtitle {
    font-size: 11px;
    color: #7a4068;
    margin-top: 4px;
    font-weight: 500;
  }

  /* ── Table ── */
  .table-wrap {
    background: #ffffff;
    border-radius: 12px;
    border: 1px solid rgba(209, 0, 113, 0.12);
    overflow: hidden;
    box-shadow: 0 1px 3px rgba(209, 0, 113, 0.04);
  }

  table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0;
    font-size: 10px;
  }

  thead {
    background: linear-gradient(135deg, #1a0010 0%, #2d0020 100%);
  }

  th {
    color: #ffd313;
    padding: 10px 14px;
    text-align: left;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.8px;
    font-size: 9px;
    border-bottom: 2px solid rgba(209, 0, 113, 0.2);
  }

  th:first-child {
    border-top-left-radius: 12px;
  }

  th:last-child {
    border-top-right-radius: 12px;
  }

  td {
    padding: 9px 14px;
    border-bottom: 1px solid rgba(209, 0, 113, 0.08);
    color: #1a0010;
    font-weight: 500;
    line-height: 1.4;
  }

  tr:last-child td {
    border-bottom: none;
  }

  tr:nth-child(even) td {
    background: rgba(252, 228, 243, 0.35);
  }

  tr:hover td {
    background: rgba(252, 228, 243, 0.6);
  }

  /* ── Footer ── */
  .footer {
    margin-top: 24px;
    padding-top: 16px;
    border-top: 1px solid rgba(209, 0, 113, 0.12);
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 9px;
    color: #a06888;
  }

  .footer-brand {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
  }

  .footer-brand .dot {
    width: 6px;
    height: 6px;
    background: linear-gradient(135deg, #d10071, #ff9100);
    border-radius: 50%;
    display: inline-block;
  }

  .footer-right {
    text-align: right;
    line-height: 1.5;
  }

  .footer-right strong {
    color: #d10071;
    font-weight: 700;
  }

  /* ── Badge / Stats ── */
  .stats-bar {
    display: flex;
    gap: 16px;
    margin-bottom: 20px;
  }

  .stat-item {
    background: #ffffff;
    border: 1px solid rgba(209, 0, 113, 0.12);
    border-radius: 8px;
    padding: 10px 16px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .stat-item .stat-label {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.8px;
    color: #7a4068;
    font-weight: 600;
  }

  .stat-item .stat-value {
    font-size: 16px;
    font-weight: 800;
    color: #d10071;
  }

  /* ── Page break helper ── */
  @media print {
    tr { page-break-inside: avoid; }
    thead { display: table-header-group; }
  }
</style>
</head>
<body>

  <!-- Header -->
  <div class="header">
    <div class="brand">
      <!-- <div class="brand-logo">W</div> -->
      <div class="brand-text">
        <h1>WAFA INDONESIA</h1>
        <div class="tagline">Belajar Al-Qur'an Metode Otak Kanan</div>
      </div>
    </div>
    <div class="meta">
      <div class="meta-label">Dokumen</div>
      <div class="meta-value">{{.Title}}</div>
      <div style="margin-top: 6px;" class="meta-label">Tanggal Cetak</div>
      <div class="meta-value">{{.Date}}</div>
    </div>
  </div>

  <!-- Title -->
  <div class="doc-title">
    <h2>{{.Title}}</h2>
    <div class="subtitle">Yayasan Syafa'atul Qur'an Indonesia (YAQIN)</div>
  </div>

  <!-- Stats -->
  <div class="stats-bar">
    <div class="stat-item">
      <span class="stat-label">Total Data</span>
      <span class="stat-value">{{.TotalData}}</span>
    </div>
    <div class="stat-item">
      <span class="stat-label">Periode</span>
      <span class="stat-value" style="font-size: 13px; color: #1a0010;">{{.Date}}</span>
    </div>
  </div>

  <!-- Table -->
  <div class="table-wrap">
    <table>
      <thead>
        <tr>{{range .Headers}}<th>{{.}}</th>{{end}}</tr>
      </thead>
      <tbody>
        {{range .Rows}}<tr>{{range .}}<td>{{.}}</td>{{end}}</tr>{{end}}
      </tbody>
    </table>
  </div>

  <!-- Footer -->
  <div class="footer">
    <div class="footer-brand">
      <span class="dot"></span>
      <span>HRIS Wafa Indonesia</span>
    </div>
    <div class="footer-right">
      <div>Dicetak pada <strong>{{.Date}}</strong></div>
      <div>wafaindonesia.or.id</div>
    </div>
  </div>

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
