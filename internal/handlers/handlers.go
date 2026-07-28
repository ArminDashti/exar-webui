package handlers

import (
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/armin/expenses/backend/internal/database"
	"github.com/armin/expenses/backend/internal/models"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *database.DB
}

func New(db *database.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) ListPersons(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, name FROM persons ORDER BY id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list persons"})
		return
	}
	defer rows.Close()

	persons := make([]models.Person, 0)
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read persons"})
			return
		}
		persons = append(persons, p)
	}

	c.JSON(http.StatusOK, persons)
}

func (h *Handler) ListShops(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, name FROM shops ORDER BY name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list shops"})
		return
	}
	defer rows.Close()

	shops := make([]models.Shop, 0)
	for rows.Next() {
		var s models.Shop
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read shops"})
			return
		}
		shops = append(shops, s)
	}

	c.JSON(http.StatusOK, shops)
}

func (h *Handler) CreateShop(c *gin.Context) {
	var req models.CreateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	result, err := h.db.Exec(`INSERT INTO shops (name) VALUES (?)`, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "shop already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create shop"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, models.Shop{ID: int(id), Name: name})
}

func (h *Handler) UpdateShop(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop id"})
		return
	}

	var req models.UpdateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	result, err := h.db.Exec(`UPDATE shops SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "shop already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update shop"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	c.JSON(http.StatusOK, models.Shop{ID: id, Name: name})
}

func (h *Handler) DeleteShop(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop id"})
		return
	}

	var inUse int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM invoices WHERE shop_id = ?`, id).Scan(&inUse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check shop usage"})
		return
	}
	if inUse > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "shop is used by invoices"})
		return
	}

	result, err := h.db.Exec(`DELETE FROM shops WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete shop"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListItems(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q != "" && len([]rune(q)) < 3 {
		c.JSON(http.StatusOK, []models.Item{})
		return
	}

	var (
		rows *sql.Rows
		err  error
	)
	if q != "" {
		rows, err = h.db.Query(
			`SELECT id, name FROM items WHERE name LIKE ? COLLATE NOCASE ORDER BY name`,
			"%"+q+"%",
		)
	} else {
		rows, err = h.db.Query(`SELECT id, name FROM items ORDER BY name`)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list items"})
		return
	}
	defer rows.Close()

	items := make([]models.Item, 0)
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read items"})
			return
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateItem(c *gin.Context) {
	var req models.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	result, err := h.db.Exec(`INSERT INTO items (name) VALUES (?)`, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "item already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create item"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, models.Item{ID: int(id), Name: name})
}

func (h *Handler) UpdateItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	var req models.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	result, err := h.db.Exec(`UPDATE items SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "item already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update item"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.JSON(http.StatusOK, models.Item{ID: id, Name: name})
}

func (h *Handler) DeleteItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	result, err := h.db.Exec(`DELETE FROM items WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete item"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetStats(c *gin.Context) {
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	where := []string{"1=1"}
	args := make([]any, 0)
	if fromDate != "" {
		where = append(where, "i.date >= ?")
		args = append(args, fromDate)
	}
	if toDate != "" {
		where = append(where, "i.date <= ?")
		args = append(args, toDate)
	}
	whereSQL := strings.Join(where, " AND ")

	var total float64
	totalQuery := `SELECT COALESCE(SUM(i.total), 0) FROM invoices i WHERE ` + whereSQL
	if err := h.db.QueryRow(totalQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute total"})
		return
	}

	byPerson := make([]models.PersonTotal, 0)
	personQuery := `
		SELECT p.id, p.name, COALESCE(SUM(i.total * sh.share), 0)
		FROM persons p
		LEFT JOIN invoice_shares sh ON sh.person_id = p.id
		LEFT JOIN invoices i ON i.id = sh.invoice_id AND ` + whereSQL + `
		GROUP BY p.id, p.name
		ORDER BY p.id`
	personRows, err := h.db.Query(personQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute person stats"})
		return
	}
	defer personRows.Close()
	for personRows.Next() {
		var row models.PersonTotal
		if err := personRows.Scan(&row.PersonID, &row.PersonName, &row.Total); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read person stats"})
			return
		}
		byPerson = append(byPerson, row)
	}

	byShop := make([]models.ShopTotal, 0)
	shopQuery := `
		SELECT s.id, s.name, SUM(i.total) AS total
		FROM invoices i
		JOIN shops s ON s.id = i.shop_id
		WHERE ` + whereSQL + `
		GROUP BY s.id, s.name
		ORDER BY total DESC`
	shopRows, err := h.db.Query(shopQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute shop stats"})
		return
	}
	defer shopRows.Close()
	for shopRows.Next() {
		var row models.ShopTotal
		if err := shopRows.Scan(&row.ShopID, &row.ShopName, &row.Total); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read shop stats"})
			return
		}
		byShop = append(byShop, row)
	}

	c.JSON(http.StatusOK, models.Stats{
		Total:    total,
		ByPerson: byPerson,
		ByShop:   byShop,
	})
}

func (h *Handler) ListInvoices(c *gin.Context) {
	query := `
		SELECT i.id, i.person_id, p.name, i.shop_id, s.name, i.date, i.total
		FROM invoices i
		JOIN persons p ON p.id = i.person_id
		JOIN shops s ON s.id = i.shop_id
		WHERE 1=1
	`
	args := []any{}

	if personID := c.Query("person_id"); personID != "" {
		query += ` AND i.person_id = ?`
		args = append(args, personID)
	}
	if from := c.Query("from_date"); from != "" {
		query += ` AND i.date >= ?`
		args = append(args, from)
	}
	if to := c.Query("to_date"); to != "" {
		query += ` AND i.date <= ?`
		args = append(args, to)
	}

	query += ` ORDER BY i.date DESC, i.id DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list invoices"})
		return
	}
	defer rows.Close()

	invoices := make([]models.Invoice, 0)
	var invoiceIDs []int
	for rows.Next() {
		var inv models.Invoice
		if err := rows.Scan(&inv.ID, &inv.PersonID, &inv.PersonName, &inv.ShopID, &inv.ShopName, &inv.Date, &inv.Total); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read invoices"})
			return
		}
		inv.Items = []models.InvoiceItem{}
		inv.Shares = []models.InvoiceShare{}
		invoices = append(invoices, inv)
		invoiceIDs = append(invoiceIDs, inv.ID)
	}

	if len(invoiceIDs) > 0 {
		itemsByInvoice, err := h.loadItemsForInvoices(invoiceIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load invoice items"})
			return
		}
		sharesByInvoice, err := h.loadSharesForInvoices(invoiceIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load invoice shares"})
			return
		}
		for i := range invoices {
			if items, ok := itemsByInvoice[invoices[i].ID]; ok {
				invoices[i].Items = items
			}
			if shares, ok := sharesByInvoice[invoices[i].ID]; ok {
				invoices[i].Shares = shares
			}
		}
	}

	c.JSON(http.StatusOK, invoices)
}

func (h *Handler) loadItemsForInvoices(invoiceIDs []int) (map[int][]models.InvoiceItem, error) {
	placeholders := make([]string, len(invoiceIDs))
	args := make([]any, len(invoiceIDs))
	for i, id := range invoiceIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT id, invoice_id, description, amount, quantity
		FROM invoice_items WHERE invoice_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY id`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]models.InvoiceItem)
	for rows.Next() {
		var item models.InvoiceItem
		if err := rows.Scan(&item.ID, &item.InvoiceID, &item.Description, &item.Amount, &item.Quantity); err != nil {
			return nil, err
		}
		result[item.InvoiceID] = append(result[item.InvoiceID], item)
	}

	return result, nil
}

func (h *Handler) loadSharesForInvoices(invoiceIDs []int) (map[int][]models.InvoiceShare, error) {
	placeholders := make([]string, len(invoiceIDs))
	args := make([]any, len(invoiceIDs))
	for i, id := range invoiceIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT sh.invoice_id, sh.person_id, p.name, sh.share
		FROM invoice_shares sh
		JOIN persons p ON p.id = sh.person_id
		WHERE sh.invoice_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY sh.person_id`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]models.InvoiceShare)
	for rows.Next() {
		var invoiceID int
		var share models.InvoiceShare
		if err := rows.Scan(&invoiceID, &share.PersonID, &share.PersonName, &share.Share); err != nil {
			return nil, err
		}
		result[invoiceID] = append(result[invoiceID], share)
	}

	return result, nil
}

func (h *Handler) GetInvoice(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice id"})
		return
	}

	inv, err := h.fetchInvoice(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get invoice"})
		return
	}

	c.JSON(http.StatusOK, inv)
}

func (h *Handler) fetchInvoice(id int) (models.Invoice, error) {
	var inv models.Invoice
	err := h.db.QueryRow(`
		SELECT i.id, i.person_id, p.name, i.shop_id, s.name, i.date, i.total
		FROM invoices i
		JOIN persons p ON p.id = i.person_id
		JOIN shops s ON s.id = i.shop_id
		WHERE i.id = ?`, id,
	).Scan(&inv.ID, &inv.PersonID, &inv.PersonName, &inv.ShopID, &inv.ShopName, &inv.Date, &inv.Total)
	if err != nil {
		return inv, err
	}

	itemsByInvoice, err := h.loadItemsForInvoices([]int{id})
	if err != nil {
		return inv, err
	}
	inv.Items = itemsByInvoice[id]
	if inv.Items == nil {
		inv.Items = []models.InvoiceItem{}
	}

	sharesByInvoice, err := h.loadSharesForInvoices([]int{id})
	if err != nil {
		return inv, err
	}
	inv.Shares = sharesByInvoice[id]
	if inv.Shares == nil {
		inv.Shares = []models.InvoiceShare{}
	}

	return inv, nil
}

func validateShares(shares []models.InvoiceShareInput) error {
	seen := make(map[int]bool, len(shares))
	var sum float64
	for _, s := range shares {
		if s.PersonID != 1 && s.PersonID != 2 {
			return errors.New("share person_id must be 1 or 2")
		}
		if s.Share < 0 {
			return errors.New("share must be >= 0")
		}
		if seen[s.PersonID] {
			return errors.New("duplicate person in shares")
		}
		seen[s.PersonID] = true
		sum += s.Share
	}

	for _, id := range []int{1, 2} {
		if !seen[id] {
			return errors.New("shares must include every person")
		}
	}

	if math.Abs(sum-1) > 0.001 {
		return errors.New("shares must sum to 1")
	}
	return nil
}

func insertShares(tx *sql.Tx, invoiceID int64, shares []models.InvoiceShareInput) error {
	for _, s := range shares {
		_, err := tx.Exec(
			`INSERT INTO invoice_shares (invoice_id, person_id, share) VALUES (?, ?, ?)`,
			invoiceID, s.PersonID, s.Share,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) CreateInvoice(c *gin.Context) {
	var req models.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PersonID != 1 && req.PersonID != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "person_id must be 1 or 2"})
		return
	}

	if err := validateShares(req.Shares); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var shopExists int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM shops WHERE id = ?`, req.ShopID).Scan(&shopExists); err != nil || shopExists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop_id"})
		return
	}

	var total float64
	for _, item := range req.Items {
		total += item.Amount
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO invoices (person_id, shop_id, date, total) VALUES (?, ?, ?, ?)`,
		req.PersonID, req.ShopID, req.Date, total,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invoice"})
		return
	}

	invoiceID, _ := result.LastInsertId()

	for _, item := range req.Items {
		_, err := tx.Exec(
			`INSERT INTO invoice_items (invoice_id, description, amount, quantity) VALUES (?, ?, ?, ?)`,
			invoiceID, strings.TrimSpace(item.Description), item.Amount, 1,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invoice items"})
			return
		}
	}

	if err := insertShares(tx, invoiceID, req.Shares); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invoice shares"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save invoice"})
		return
	}

	inv, err := h.fetchInvoice(int(invoiceID))
	if err != nil {
		c.JSON(http.StatusCreated, gin.H{"id": invoiceID})
		return
	}

	c.JSON(http.StatusCreated, inv)
}

func (h *Handler) UpdateInvoice(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice id"})
		return
	}

	var req models.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PersonID != 1 && req.PersonID != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "person_id must be 1 or 2"})
		return
	}

	if err := validateShares(req.Shares); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var shopExists int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM shops WHERE id = ?`, req.ShopID).Scan(&shopExists); err != nil || shopExists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop_id"})
		return
	}

	var exists int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM invoices WHERE id = ?`, id).Scan(&exists); err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		return
	}

	var total float64
	for _, item := range req.Items {
		total += item.Amount
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE invoices SET person_id = ?, shop_id = ?, date = ?, total = ? WHERE id = ?`,
		req.PersonID, req.ShopID, req.Date, total, id,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update invoice"})
		return
	}

	if _, err := tx.Exec(`DELETE FROM invoice_items WHERE invoice_id = ?`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to replace invoice items"})
		return
	}

	for _, item := range req.Items {
		_, err := tx.Exec(
			`INSERT INTO invoice_items (invoice_id, description, amount, quantity) VALUES (?, ?, ?, ?)`,
			id, strings.TrimSpace(item.Description), item.Amount, 1,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invoice items"})
			return
		}
	}

	if _, err := tx.Exec(`DELETE FROM invoice_shares WHERE invoice_id = ?`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to replace invoice shares"})
		return
	}

	if err := insertShares(tx, int64(id), req.Shares); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invoice shares"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save invoice"})
		return
	}

	inv, err := h.fetchInvoice(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id})
		return
	}

	c.JSON(http.StatusOK, inv)
}

func (h *Handler) DeleteInvoice(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice id"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM invoice_shares WHERE invoice_id = ?`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete invoice shares"})
		return
	}

	if _, err := tx.Exec(`DELETE FROM invoice_items WHERE invoice_id = ?`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete invoice items"})
		return
	}

	result, err := tx.Exec(`DELETE FROM invoices WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete invoice"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit delete"})
		return
	}

	c.Status(http.StatusNoContent)
}
