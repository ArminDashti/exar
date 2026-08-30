package web // sets the Go package name to web

import ( // starts the import list for required packages
	"bytes" // executes this step of the current logic
	"crypto/subtle" // executes this step of the current logic
	"encoding/json" // executes this step of the current logic
	"errors" // executes this step of the current logic
	"fmt" // executes this step of the current logic
	"net/http" // executes this step of the current logic
	"sort" // executes this step of the current logic
	"strconv" // executes this step of the current logic
	"strings" // executes this step of the current logic
	"time" // executes this step of the current logic

	"expenses/internal/config" // executes this step of the current logic
	"expenses/internal/database" // executes this step of the current logic
) // ends the current grouped declaration

func RegisterRoutes(mux *http.ServeMux) { // defines RegisterRoutes, which handles this unit of behavior
	mux.HandleFunc("/", handleIndex) // executes this step of the current logic
	mux.HandleFunc("/expenses", handleExpenses) // executes this step of the current logic
	mux.HandleFunc("/expenses/api/v1/expenses-list", withAPIToken(handleAPIExpensesList)) // executes this step of the current logic
	mux.HandleFunc("/expenses/api/v1/add-expense", withAPIToken(handleAPIAddExpense)) // executes this step of the current logic
	mux.HandleFunc("/expenses/api/v1/delete-expense", withAPIToken(handleAPIDeleteExpense)) // executes this step of the current logic
} // closes the current block scope

func withAPIToken(next http.HandlerFunc) http.HandlerFunc { // defines withAPIToken, which handles this unit of behavior
	return func(w http.ResponseWriter, r *http.Request) { // returns the computed values to the caller
		expectedToken := strings.TrimSpace(config.APIToken()) // initializes expectedToken with the result of this expression
		if expectedToken == "" { // checks this condition before continuing
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server token is not configured"}) // executes this step of the current logic
			return // returns from the current function
		} // closes the current block scope

		providedToken := strings.TrimSpace(r.Header.Get("X-API-Token")) // initializes providedToken with the result of this expression
		if providedToken == "" { // checks this condition before continuing
			providedToken = strings.TrimSpace(r.URL.Query().Get("token")) // updates providedToken with a new value
		} // closes the current block scope
		if subtle.ConstantTimeCompare([]byte(providedToken), []byte(expectedToken)) != 1 { // checks this condition before continuing
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid api token"}) // executes this step of the current logic
			return // returns from the current function
		} // closes the current block scope

		next(w, r) // executes this step of the current logic
	} // closes the current block scope
} // closes the current block scope

func handleIndex(w http.ResponseWriter, r *http.Request) { // defines handleIndex, which handles this unit of behavior
	if r.URL.Path != "/" { // checks this condition before continuing
		http.NotFound(w, r) // calls an HTTP helper to build or process this response
		return // returns from the current function
	} // closes the current block scope
	if r.Method != http.MethodGet { // checks this condition before continuing
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // calls an HTTP helper to build or process this response
		return // returns from the current function
	} // closes the current block scope
	renderForm(w, http.StatusOK, pageData{Values: map[string]string{}}) // executes this step of the current logic
} // closes the current block scope

func handleExpenses(w http.ResponseWriter, r *http.Request) { // defines handleExpenses, which handles this unit of behavior
	if r.URL.Path != "/expenses" { // checks this condition before continuing
		http.NotFound(w, r) // calls an HTTP helper to build or process this response
		return // returns from the current function
	} // closes the current block scope
	if r.Method != http.MethodPost { // checks this condition before continuing
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // calls an HTTP helper to build or process this response
		return // returns from the current function
	} // closes the current block scope
	if err := r.ParseForm(); err != nil { // checks this condition before continuing
		renderForm(w, http.StatusBadRequest, pageData{ // opens a new block for the following statements
			Errors: []string{"Could not parse the submitted form."}, // executes this step of the current logic
			Values: map[string]string{}, // executes this step of the current logic
		}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope

	form := map[string]string{ // opens a new block for the following statements
		"id":            strings.TrimSpace(r.FormValue("id")), // executes this step of the current logic
		"store_id":      strings.TrimSpace(r.FormValue("store_id")), // executes this step of the current logic
		"purchase":      strings.TrimSpace(r.FormValue("purchase")), // executes this step of the current logic
		"purchase_time": strings.TrimSpace(r.FormValue("purchase_time")), // executes this step of the current logic
		"owner":         strings.TrimSpace(r.FormValue("owner")), // executes this step of the current logic
		"inserted_at":   strings.TrimSpace(r.FormValue("inserted_at")), // executes this step of the current logic
	} // closes the current block scope

	errors := validate(form) // initializes errors with the result of this expression
	if len(errors) > 0 { // checks this condition before continuing
		renderForm(w, http.StatusOK, pageData{Errors: errors, Values: form}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope

	if err := database.InsertExpense(form); err != nil { // checks this condition before continuing
		renderForm(w, http.StatusOK, pageData{ // opens a new block for the following statements
			Errors: []string{err.Error()}, // executes this step of the current logic
			Values: form, // executes this step of the current logic
		}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope

	renderForm(w, http.StatusOK, pageData{ // opens a new block for the following statements
		Success: "Expense inserted successfully.", // executes this step of the current logic
		Values:  map[string]string{}, // executes this step of the current logic
	}) // executes this step of the current logic
} // closes the current block scope

type expenseListItem struct { // declares the expenseListItem struct type
	PurchaseDate string `json:"purchase-date"` // executes this step of the current logic
	Store        string `json:"store"` // executes this step of the current logic
	Purchase     string `json:"purchase"` // executes this step of the current logic
} // closes the current block scope

type orderedExpenseList []expenseListItem // executes this step of the current logic

func (records orderedExpenseList) MarshalJSON() ([]byte, error) { // defines MarshalJSON, which handles this unit of behavior
	var buffer bytes.Buffer // declares a package-level variable
	buffer.WriteByte('{') // executes this step of the current logic

	for index, record := range records { // iterates through items while this loop condition holds
		if index > 0 { // checks this condition before continuing
			buffer.WriteByte(',') // executes this step of the current logic
		} // closes the current block scope

		dateJSON, err := json.Marshal(record.PurchaseDate) // initializes dateJSON, err with the result of this expression
		if err != nil { // checks this condition before continuing
			return nil, err // returns the computed values to the caller
		} // closes the current block scope
		valueJSON, err := json.Marshal(map[string]string{ // opens a new block for the following statements
			"store":    record.Store, // executes this step of the current logic
			"purchase": record.Purchase, // executes this step of the current logic
		}) // executes this step of the current logic
		if err != nil { // checks this condition before continuing
			return nil, err // returns the computed values to the caller
		} // closes the current block scope

		buffer.Write(dateJSON) // executes this step of the current logic
		buffer.WriteByte(':') // executes this step of the current logic
		buffer.Write(valueJSON) // executes this step of the current logic
	} // closes the current block scope

	buffer.WriteByte('}') // executes this step of the current logic
	return buffer.Bytes(), nil // returns the computed values to the caller
} // closes the current block scope

func handleAPIExpensesList(w http.ResponseWriter, r *http.Request) { // defines handleAPIExpensesList, which handles this unit of behavior
	if r.Method != http.MethodGet { // checks this condition before continuing
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope
	query := r.URL.Query() // initializes query with the result of this expression

	owner := strings.TrimSpace(query.Get("owner")) // initializes owner with the result of this expression
	fromDate := strings.TrimSpace(query.Get("from-date")) // initializes fromDate with the result of this expression
	toDate := strings.TrimSpace(query.Get("to-date")) // initializes toDate with the result of this expression

	if owner != "" { // checks this condition before continuing
		if _, err := strconv.Atoi(owner); err != nil { // checks this condition before continuing
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "owner must be a whole number"}) // executes this step of the current logic
			return // returns from the current function
		} // closes the current block scope
	} // closes the current block scope
	if fromDate != "" && !validDate(fromDate) { // checks this condition before continuing
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "from-date must use YYYY-MM-DD (Shamsi or Gregorian)"}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope
	if toDate != "" && !validDate(toDate) { // checks this condition before continuing
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to-date must use YYYY-MM-DD (Shamsi or Gregorian)"}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope
	if fromDate != "" { // checks this condition before continuing
		converted, err := normalizeToGregorianDate(fromDate) // initializes converted, err with the result of this expression
		if err != nil { // checks this condition before continuing
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()}) // executes this step of the current logic
			return // returns from the current function
		} // closes the current block scope
		fromDate = converted // updates fromDate with a new value
	} // closes the current block scope
	if toDate != "" { // checks this condition before continuing
		converted, err := normalizeToGregorianDate(toDate) // initializes converted, err with the result of this expression
		if err != nil { // checks this condition before continuing
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()}) // executes this step of the current logic
			return // returns from the current function
		} // closes the current block scope
		toDate = converted // updates toDate with a new value
	} // closes the current block scope

	records, err := database.GetExpenses(database.ExpenseFilters{ // opens a new block for the following statements
		Owner:    owner, // executes this step of the current logic
		FromDate: fromDate, // executes this step of the current logic
		ToDate:   toDate, // executes this step of the current logic
	}) // executes this step of the current logic
	if err != nil { // checks this condition before continuing
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope

	listItems := make([]expenseListItem, 0, len(records)) // initializes listItems with the result of this expression
	for _, record := range records { // iterates through items while this loop condition holds
		purchaseDate, ok := record["purchase-date"] // initializes purchaseDate, ok with the result of this expression
		if !ok { // checks this condition before continuing
			purchaseDate = strings.TrimSpace(record["purchase_time"]) // updates purchaseDate with a new value
		} // closes the current block scope
		purchaseDate = strings.TrimSpace(purchaseDate) // updates purchaseDate with a new value
		if purchaseDate == "" { // checks this condition before continuing
			continue // skips to the next loop iteration
		} // closes the current block scope

		store := strings.TrimSpace(record["store"]) // initializes store with the result of this expression
		if store == "" { // checks this condition before continuing
			store = strings.TrimSpace(record["store_id"]) // updates store with a new value
		} // closes the current block scope
		purchase := strings.TrimSpace(record["purchase"]) // initializes purchase with the result of this expression

		listItems = append(listItems, expenseListItem{ // opens a new block for the following statements
			PurchaseDate: purchaseDate, // executes this step of the current logic
			Store:        store, // executes this step of the current logic
			Purchase:     purchase, // executes this step of the current logic
		}) // executes this step of the current logic
	} // closes the current block scope

	sort.Slice(listItems, func(i, j int) bool { // opens a new block for the following statements
		return listItems[i].PurchaseDate < listItems[j].PurchaseDate // returns the computed values to the caller
	}) // executes this step of the current logic

	writeJSON(w, http.StatusOK, orderedExpenseList(listItems)) // executes this step of the current logic
} // closes the current block scope

func handleAPIAddExpense(w http.ResponseWriter, r *http.Request) { // defines handleAPIAddExpense, which handles this unit of behavior
	if r.Method != http.MethodPost { // checks this condition before continuing
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope

	payload, err := readExpensePayload(r) // initializes payload, err with the result of this expression
	if err != nil { // checks this condition before continuing
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope

	if !config.HasDatabaseConfig() { // checks this condition before continuing
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "database config is missing"}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope
	if _, convErr := strconv.Atoi(payload.Owner); convErr != nil { // checks this condition before continuing
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "owner must be a whole number"}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope
	if !validDateTime(payload.PurchaseDate) { // checks this condition before continuing
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "purchase-date must use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS (Shamsi or Gregorian)"}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope
	convertedDate, convErr := normalizeToGregorianDateTime(payload.PurchaseDate) // initializes convertedDate, convErr with the result of this expression
	if convErr != nil { // checks this condition before continuing
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": convErr.Error()}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope
	payload.PurchaseDate = convertedDate // updates payload.PurchaseDate with a new value

	if dbErr := database.AddExpense(payload.Owner, payload.Store, payload.Purchase, payload.PurchaseDate); dbErr != nil { // checks this condition before continuing
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": dbErr.Error()}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope

	writeJSON(w, http.StatusCreated, map[string]string{"status": "expense inserted"}) // executes this step of the current logic
} // closes the current block scope

func handleAPIDeleteExpense(w http.ResponseWriter, r *http.Request) { // defines handleAPIDeleteExpense, which handles this unit of behavior
	if r.Method != http.MethodDelete { // checks this condition before continuing
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope

	purchase := strings.TrimSpace(r.URL.Query().Get("purchase")) // initializes purchase with the result of this expression
	purchaseDate := strings.TrimSpace(r.URL.Query().Get("purchase-date")) // initializes purchaseDate with the result of this expression
	if purchase == "" || purchaseDate == "" { // checks this condition before continuing
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "purchase and purchase-date are required"}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope
	if !validDate(purchaseDate) { // checks this condition before continuing
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "purchase-date must use YYYY-MM-DD"}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope

	deleted, err := database.DeleteExpense(purchase, purchaseDate) // initializes deleted, err with the result of this expression
	if err != nil { // checks this condition before continuing
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope
	if deleted == 0 { // checks this condition before continuing
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "expense not found"}) // executes this step of the current logic
		return // returns from the current function
	} // closes the current block scope

	writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted}) // executes this step of the current logic
} // closes the current block scope

func validate(form map[string]string) []string { // defines validate, which handles this unit of behavior
	required := map[string]string{ // opens a new block for the following statements
		"id":            "ID", // executes this step of the current logic
		"store_id":      "Store", // executes this step of the current logic
		"purchase":      "Purchase", // executes this step of the current logic
		"purchase_time": "Purchase time", // executes this step of the current logic
		"owner":         "Owner", // executes this step of the current logic
	} // closes the current block scope

	var errors []string // declares a package-level variable
	for field, label := range required { // iterates through items while this loop condition holds
		if form[field] == "" { // checks this condition before continuing
			errors = append(errors, label+" is required.") // updates errors with a new value
		} // closes the current block scope
	} // closes the current block scope

	if !config.HasDatabaseConfig() { // checks this condition before continuing
		errors = append(errors, "Set DATABASE_URL or PGHOST/PGPORT/PGDATABASE/PGUSER/PGPASSWORD before submitting.") // updates errors with a new value
	} // closes the current block scope
	if form["owner"] != "" { // checks this condition before continuing
		if _, err := strconv.Atoi(form["owner"]); err != nil { // checks this condition before continuing
			errors = append(errors, "Owner must be a whole number.") // updates errors with a new value
		} // closes the current block scope
	} // closes the current block scope

	return errors // returns the computed values to the caller
} // closes the current block scope

type addExpensePayload struct { // declares the addExpensePayload struct type
	Owner        string `json:"owner"` // executes this step of the current logic
	Store        string `json:"store"` // executes this step of the current logic
	Purchase     string `json:"purchase"` // executes this step of the current logic
	PurchaseDate string `json:"purchase-date"` // executes this step of the current logic
} // closes the current block scope

func readExpensePayload(r *http.Request) (addExpensePayload, error) { // defines readExpensePayload, which handles this unit of behavior
	var payload addExpensePayload // declares a package-level variable
	contentType := r.Header.Get("Content-Type") // initializes contentType with the result of this expression

	if strings.HasPrefix(contentType, "application/json") { // checks this condition before continuing
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { // checks this condition before continuing
			return payload, err // returns the computed values to the caller
		} // closes the current block scope
	} else { // switches from the previous branch to an alternate path
		if err := r.ParseForm(); err != nil { // checks this condition before continuing
			return payload, err // returns the computed values to the caller
		} // closes the current block scope
		payload.Owner = r.FormValue("owner") // updates payload.Owner with a new value
		payload.Store = r.FormValue("store") // updates payload.Store with a new value
		payload.Purchase = r.FormValue("purchase") // updates payload.Purchase with a new value
		payload.PurchaseDate = r.FormValue("purchase-date") // updates payload.PurchaseDate with a new value
	} // closes the current block scope

	payload.Owner = strings.TrimSpace(payload.Owner) // updates payload.Owner with a new value
	payload.Store = strings.TrimSpace(payload.Store) // updates payload.Store with a new value
	payload.Purchase = strings.TrimSpace(payload.Purchase) // updates payload.Purchase with a new value
	payload.PurchaseDate = strings.TrimSpace(payload.PurchaseDate) // updates payload.PurchaseDate with a new value

	if payload.Owner == "" || payload.Store == "" || payload.Purchase == "" || payload.PurchaseDate == "" { // checks this condition before continuing
		return payload, errors.New("owner, store, purchase, and purchase-date are required") // returns the computed values to the caller
	} // closes the current block scope

	return payload, nil // returns the computed values to the caller
} // closes the current block scope

func validDate(value string) bool { // defines validDate, which handles this unit of behavior
	_, err := normalizeToGregorianDate(value) // initializes _, err with the result of this expression
	return err == nil // returns the computed values to the caller
} // closes the current block scope

func validDateTime(value string) bool { // defines validDateTime, which handles this unit of behavior
	_, err := normalizeToGregorianDateTime(value) // initializes _, err with the result of this expression
	return err == nil // returns the computed values to the caller
} // closes the current block scope

func writeJSON(w http.ResponseWriter, status int, payload any) { // defines writeJSON, which handles this unit of behavior
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // updates w.Header().Set("Content-Type", "application/json; charset with a new value
	w.WriteHeader(status) // executes this step of the current logic
	_ = json.NewEncoder(w).Encode(payload) // updates _ with a new value
} // closes the current block scope

func normalizeToGregorianDateTime(value string) (string, error) { // defines normalizeToGregorianDateTime, which handles this unit of behavior
	value = strings.TrimSpace(value) // updates value with a new value
	if value == "" { // checks this condition before continuing
		return "", errors.New("date value is empty") // returns the computed values to the caller
	} // closes the current block scope

	parts := strings.Split(value, " ") // initializes parts with the result of this expression
	if len(parts) > 2 { // checks this condition before continuing
		return "", errors.New("date value must use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS") // returns the computed values to the caller
	} // closes the current block scope

	gregDate, err := normalizeToGregorianDate(parts[0]) // initializes gregDate, err with the result of this expression
	if err != nil { // checks this condition before continuing
		return "", err // returns the computed values to the caller
	} // closes the current block scope
	if len(parts) == 1 { // checks this condition before continuing
		return gregDate, nil // returns the computed values to the caller
	} // closes the current block scope

	if _, err := time.Parse("15:04:05", parts[1]); err != nil { // checks this condition before continuing
		return "", errors.New("time must use HH:MM:SS") // returns the computed values to the caller
	} // closes the current block scope
	return gregDate + " " + parts[1], nil // returns the computed values to the caller
} // closes the current block scope

func normalizeToGregorianDate(value string) (string, error) { // defines normalizeToGregorianDate, which handles this unit of behavior
	value = strings.TrimSpace(value) // updates value with a new value
	parsed, err := time.Parse("2006-01-02", value) // initializes parsed, err with the result of this expression
	if err != nil { // checks this condition before continuing
		return "", errors.New("date must use YYYY-MM-DD") // returns the computed values to the caller
	} // closes the current block scope

	year := parsed.Year() // initializes year with the result of this expression
	// Treat year range 1300..1699 as Jalali input (for example 1404-05-06).
	if year >= 1300 && year < 1700 { // checks this condition before continuing
		gy, gm, gd, convErr := jalaliToGregorian(year, int(parsed.Month()), parsed.Day()) // initializes gy, gm, gd, convErr with the result of this expression
		if convErr != nil { // checks this condition before continuing
			return "", convErr // returns the computed values to the caller
		} // closes the current block scope
		return fmt.Sprintf("%04d-%02d-%02d", gy, gm, gd), nil // returns the computed values to the caller
	} // closes the current block scope

	return parsed.Format("2006-01-02"), nil // returns the computed values to the caller
} // closes the current block scope

func jalaliToGregorian(jy, jm, jd int) (int, int, int, error) { // defines jalaliToGregorian, which handles this unit of behavior
	if jy < 1 || jm < 1 || jm > 12 || jd < 1 || jd > 31 { // checks this condition before continuing
		return 0, 0, 0, errors.New("invalid Shamsi date") // returns the computed values to the caller
	} // closes the current block scope

	jDaysInMonth := []int{31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 29} // initializes jDaysInMonth with the result of this expression
	gDaysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31} // initializes gDaysInMonth with the result of this expression

	jy -= 979 // updates jy - with a new value
	jm-- // executes this step of the current logic
	jd-- // executes this step of the current logic

	jDayNo := 365*jy + (jy/33)*8 + ((jy%33)+3)/4 // initializes jDayNo with the result of this expression
	for i := 0; i < jm; i++ { // iterates through items while this loop condition holds
		jDayNo += jDaysInMonth[i] // updates jDayNo + with a new value
	} // closes the current block scope
	jDayNo += jd // updates jDayNo + with a new value

	gDayNo := jDayNo + 79 // initializes gDayNo with the result of this expression
	gy := 1600 + 400*(gDayNo/146097) // initializes gy with the result of this expression
	gDayNo %= 146097 // updates gDayNo % with a new value

	leap := true // initializes leap with the result of this expression
	if gDayNo >= 36525 { // checks this condition before continuing
		gDayNo-- // executes this step of the current logic
		gy += 100 * (gDayNo / 36524) // updates gy + with a new value
		gDayNo %= 36524 // updates gDayNo % with a new value

		if gDayNo >= 365 { // checks this condition before continuing
			gDayNo++ // executes this step of the current logic
		} else { // switches from the previous branch to an alternate path
			leap = false // updates leap with a new value
		} // closes the current block scope
	} // closes the current block scope

	gy += 4 * (gDayNo / 1461) // updates gy + with a new value
	gDayNo %= 1461 // updates gDayNo % with a new value
	if gDayNo >= 366 { // checks this condition before continuing
		leap = false // updates leap with a new value
		gDayNo-- // executes this step of the current logic
		gy += gDayNo / 365 // updates gy + with a new value
		gDayNo %= 365 // updates gDayNo % with a new value
	} // closes the current block scope

	gm := 0 // initializes gm with the result of this expression
	for ; gm < len(gDaysInMonth); gm++ { // iterates through items while this loop condition holds
		days := gDaysInMonth[gm] // initializes days with the result of this expression
		if gm == 1 && leap { // checks this condition before continuing
			days++ // executes this step of the current logic
		} // closes the current block scope
		if gDayNo < days { // checks this condition before continuing
			break // exits the current loop or switch
		} // closes the current block scope
		gDayNo -= days // updates gDayNo - with a new value
	} // closes the current block scope

	if gm >= len(gDaysInMonth) { // checks this condition before continuing
		return 0, 0, 0, errors.New("invalid Shamsi date") // returns the computed values to the caller
	} // closes the current block scope

	return gy, gm + 1, gDayNo + 1, nil // returns the computed values to the caller
} // closes the current block scope
