package web // sets the Go package name to web

import ( // starts the import list for required packages
	"fmt" // executes this step of the current logic
	"html" // executes this step of the current logic
	"net/http" // executes this step of the current logic
	"strings" // executes this step of the current logic

	"expenses/internal/config" // executes this step of the current logic
) // ends the current grouped declaration

func renderForm(w http.ResponseWriter, status int, data pageData) { // defines renderForm, which handles this unit of behavior
	if data.Values == nil { // checks this condition before continuing
		data.Values = map[string]string{} // updates data.Values with a new value
	} // closes the current block scope
	body := pageHTML(data) // initializes body with the result of this expression
	w.Header().Set("Content-Type", "text/html; charset=utf-8") // updates w.Header().Set("Content-Type", "text/html; charset with a new value
	w.WriteHeader(status) // executes this step of the current logic
	_, _ = w.Write([]byte(body)) // updates _, _ with a new value
} // closes the current block scope

func pageHTML(data pageData) string { // defines pageHTML, which handles this unit of behavior
	errorBlock := "" // initializes errorBlock with the result of this expression
	if len(data.Errors) > 0 { // checks this condition before continuing
		var items strings.Builder // declares a package-level variable
		for _, msg := range data.Errors { // iterates through items while this loop condition holds
			items.WriteString("<li>") // executes this step of the current logic
			items.WriteString(html.EscapeString(msg)) // executes this step of the current logic
			items.WriteString("</li>") // executes this step of the current logic
		} // closes the current block scope
		errorBlock = `<div class="card error"><ul>` + items.String() + `</ul></div>` // updates errorBlock with a new value
	} // closes the current block scope

	successBlock := "" // initializes successBlock with the result of this expression
	if data.Success != "" { // checks this condition before continuing
		successBlock = `<div class="card success">` + html.EscapeString(data.Success) + `</div>` // updates successBlock with a new value
	} // closes the current block scope

	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>%s</title>
  <style>
    :root {
      --bg: #f2ede3;
      --panel: #fffaf2;
      --ink: #1c1917;
      --accent: #0f766e;
      --accent-2: #f97316;
      --error: #b91c1c;
      --shadow: rgba(28, 25, 23, 0.14);
      font-family: Georgia, "Times New Roman", serif;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      background:
        radial-gradient(circle at top left, rgba(249, 115, 22, 0.22), transparent 30%%),
        linear-gradient(135deg, #efe3cf 0%%, var(--bg) 55%%, #d7eae4 100%%);
      color: var(--ink);
      display: grid;
      place-items: center;
      padding: 24px;
    }
    .shell {
      width: min(100%%, 960px);
      display: grid;
      grid-template-columns: 1.1fr 0.9fr;
      gap: 24px;
      align-items: start;
    }
    .hero, form {
      background: color-mix(in srgb, var(--panel) 94%%, white);
      border-radius: 28px;
      box-shadow: 0 22px 50px var(--shadow);
      padding: 28px;
      border: 1px solid rgba(28, 25, 23, 0.08);
    }
    .kicker {
      display: inline-block;
      letter-spacing: 0.18em;
      text-transform: uppercase;
      font-size: 12px;
      color: var(--accent);
      margin-bottom: 14px;
    }
    h1 { font-size: clamp(2.6rem, 5vw, 4.6rem); line-height: 0.95; margin: 0 0 16px; }
    .hero p { margin: 0; font-size: 1.05rem; line-height: 1.6; max-width: 32ch; }
    .chips { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 22px; }
    .chip {
      padding: 8px 12px;
      border-radius: 999px;
      background: rgba(15, 118, 110, 0.1);
      color: var(--accent);
      font-size: 0.92rem;
    }
    form { display: grid; gap: 14px; }
    label { display: grid; gap: 6px; font-size: 0.95rem; }
    span { font-weight: 700; }
    input {
      width: 100%%;
      border: 1px solid rgba(28, 25, 23, 0.16);
      border-radius: 14px;
      padding: 12px 14px;
      font: inherit;
      background: white;
    }
    input:focus { outline: 2px solid rgba(15, 118, 110, 0.22); border-color: var(--accent); }
    button {
      border: 0;
      border-radius: 16px;
      padding: 14px 18px;
      font: inherit;
      font-weight: 700;
      background: linear-gradient(135deg, var(--accent), #155e75);
      color: white;
      cursor: pointer;
    }
    button:hover { transform: translateY(-1px); }
    .card { border-radius: 18px; padding: 14px 16px; }
    .error { background: #fee2e2; color: var(--error); }
    .success { background: #dcfce7; color: #166534; }
    ul { margin: 0; padding-left: 20px; }
    .hint { font-size: 0.9rem; opacity: 0.8; margin: 0; }
    @media (max-width: 820px) {
      .shell { grid-template-columns: 1fr; }
      .hero, form { padding: 22px; }
      h1 { font-size: 3rem; }
    }
  </style>
</head>
<body>
  <main class="shell">
    <section class="hero">
      <div class="kicker">Expenses</div>
      <h1>Insert one clean row at a time.</h1>
      <p>Use this lightweight form to add records into the <code>expenses</code> table with the exact columns you listed, including the quoted Postgres names for <code>store-id</code>, <code>purchase-time</code>, and <code>inserted-at</code>.</p>
      <div class="chips">
        <div class="chip">Table: expenses</div>
        <div class="chip">Target: PostgreSQL</div>
        <div class="chip">Go stdlib only</div>
      </div>
    </section>
    <form method="post" action="/expenses">
      %s
      %s
      %s
      %s
      %s
      %s
      %s
      %s
      <button type="submit">Insert expense</button>
      <p class="hint">If <code>inserted_at</code> is left empty, the app inserts <code>NULL</code> so your table default can fill it in.</p>
    </form>
  </main>
</body>
</html>`,
		config.PageTitle, // executes this step of the current logic
		errorBlock, // executes this step of the current logic
		successBlock, // executes this step of the current logic
		inputField(data.Values, "id", "ID", "exp-1001", "text"), // executes this step of the current logic
		inputField(data.Values, "store_id", "Store", "amazon", "text"), // executes this step of the current logic
		inputField(data.Values, "purchase", "Purchase", "42.90", "text"), // executes this step of the current logic
		inputField(data.Values, "purchase_time", "Purchase time", "2025-01-15 14:30:00", "text"), // executes this step of the current logic
		inputField(data.Values, "owner", "Owner", "42", "number"), // executes this step of the current logic
		inputField(data.Values, "inserted_at", "Inserted at (optional)", "2025-01-15 14:35:00", "text"), // executes this step of the current logic
	) // ends the current grouped declaration
} // closes the current block scope

func inputField(values map[string]string, name string, label string, placeholder string, inputType string) string { // defines inputField, which handles this unit of behavior
	return fmt.Sprintf( // returns the computed values to the caller
		`<label>
      <span>%s</span>
      <input name="%s" type="%s" value="%s" placeholder="%s" />
    </label>`,
		html.EscapeString(label), // executes this step of the current logic
		html.EscapeString(name), // executes this step of the current logic
		html.EscapeString(inputType), // executes this step of the current logic
		html.EscapeString(values[name]), // executes this step of the current logic
		html.EscapeString(placeholder), // executes this step of the current logic
	) // ends the current grouped declaration
} // closes the current block scope
