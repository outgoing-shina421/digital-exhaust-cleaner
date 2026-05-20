// Package ui renders local review surfaces for scan results.
package ui

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"digital-exhaust-cleaner/internal/analyzer"
)

// WriteReport renders an analysis result to a standalone HTML file.
func WriteReport(path string, result analyzer.Result) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return fmt.Errorf("create report directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}
	defer file.Close()

	if err := reportTemplate.Execute(file, viewModel{Result: result, Interactive: false}); err != nil {
		return fmt.Errorf("render report: %w", err)
	}
	return nil
}

type viewModel struct {
	Result      analyzer.Result
	Interactive bool
}

func (v viewModel) TotalRecoverable() int64 {
	var total int64
	for _, group := range v.Result.DuplicateGroups {
		total += group.WastedBytes
	}
	return total
}

var reportTemplate = template.Must(template.New("report").Funcs(template.FuncMap{
	"bytes": formatBytes,
	"pct":   formatPercent,
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Digital Exhaust Cleaner</title>
  {{if .Interactive}}
  <link rel="stylesheet" href="/static/app.css?v=1.1">
  {{end}}
</head>
<body>

{{if .Interactive}}
<div class="scanning-overlay" id="overlay">
  <div class="spinner"></div>
  <p id="overlay-msg">Scanning directory…</p>
</div>
{{end}}

<header>
  <div class="header-top">
    <h1>Digital Exhaust Cleaner</h1>
    <span class="badge">Local · Private</span>
  </div>
  <div class="current-path">📁 {{.Result.Root}}</div>

  {{if .Interactive}}
  <div class="dir-picker">
    <button type="button" class="btn btn--picker" id="btn-pick" title="Choose a folder">
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
      </svg>
      Browse Folder
    </button>
    <input type="text" id="path-input" value="{{.Result.Root}}" placeholder="Or type an absolute path…" autocomplete="off" spellcheck="false">
    <button type="button" class="btn btn--primary" id="btn-scan">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
      </svg>
      Scan
    </button>
  </div>
  {{end}}
</header>

<main>
  <div class="metrics">
    <div class="metric"><span>Files scanned</span><strong>{{.Result.FilesScanned}}</strong></div>
    <div class="metric"><span>Recommendations</span><strong>{{len .Result.Recommendations}}</strong></div>
    <div class="metric"><span>Duplicate groups</span><strong>{{len .Result.DuplicateGroups}}</strong></div>
    <div class="metric"><span>Recoverable space</span><strong>{{bytes .TotalRecoverable}}</strong></div>
  </div>

  <section>
    <h2>
      Cleanup Recommendations
      <span class="score-legend">
        <span><span class="dot dot--high"></span> High (&gt;0.75)</span>
        <span><span class="dot dot--medium"></span> Medium (0.5–0.75)</span>
        <span><span class="dot dot--low"></span> Low (&lt;0.5)</span>
      </span>
    </h2>
    {{if .Result.Recommendations}}
    <div class="table-wrap">
      <table>
        <colgroup>
          <col class="col-score">
          <col class="col-cat">
          <col class="col-expl">
          <col class="col-path">
          {{if .Interactive}}<col class="col-action">{{end}}
        </colgroup>
        <thead>
          <tr>
            <th title="Confidence score 0–1. Higher means more likely to be clutter.">Score ↑</th>
            <th>Category</th>
            <th>Explanation</th>
            <th>Path</th>
            {{if .Interactive}}<th>Action</th>{{end}}
          </tr>
        </thead>
        <tbody>
          {{range .Result.Recommendations}}
          <tr>
            <td>
              {{if ge .Score 0.75}}
                <span class="score score--high">{{printf "%.2f" .Score}}</span>
              {{else if ge .Score 0.5}}
                <span class="score score--medium">{{printf "%.2f" .Score}}</span>
              {{else}}
                <span class="score score--low">{{printf "%.2f" .Score}}</span>
              {{end}}
            </td>
            <td><span class="category">{{.Category}}</span><div class="rules">{{range .Rules}}{{.}} {{end}}</div></td>
            <td>{{.Explanation}}</td>
            <td class="path-cell">{{.Path}}</td>
            {{if $.Interactive}}
            <td class="action-cell">
              <button class="btn danger" data-path="{{.Path}}" onclick="quarantine(this)">Quarantine</button>
            </td>
            {{end}}
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
    {{else}}
    <div class="empty">✓ No cleanup recommendations were generated.</div>
    {{end}}
  </section>

  <section>
    <h2>Behavior &amp; Semantic Signals</h2>
    {{if or .Result.Classifications .Result.Findings}}
    <div class="table-wrap">
      <table>
        <colgroup>
          <col class="col-cat">
          <col style="width:90px">
          <col class="col-expl">
          <col class="col-path">
        </colgroup>
        <thead>
          <tr><th>Type</th><th>Confidence</th><th>Explanation</th><th>Path</th></tr>
        </thead>
        <tbody>
          {{range .Result.Classifications}}
          <tr>
            <td class="category">{{.Label}}</td>
            <td>{{pct .Confidence}}</td>
            <td>{{.Explanation}}</td>
            <td class="path-cell">{{.Path}}</td>
          </tr>
          {{end}}
          {{range .Result.Findings}}
          <tr>
            <td class="category">{{.Pattern}}</td>
            <td>{{pct .Confidence}}</td>
            <td>{{.Explanation}}</td>
            <td class="path-cell">{{.Path}}</td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
    {{else}}
    <div class="empty">No behavioral or semantic signals were found.</div>
    {{end}}
  </section>
</main>

{{if .Interactive}}
<script src="/static/app.js?v=1.1"></script>
{{end}}
</body>
</html>`))

func formatBytes(value int64) string {
	const unit = 1024
	if value < unit {
		return fmt.Sprintf("%d B", value)
	}
	div, exp := int64(unit), 0
	for n := value / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(value)/float64(div), "KMGTPE"[exp])
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.0f%%", value*100)
}
