package users

import "fmt"

type query struct {
	query       string
	whereExist  bool
	numOfFilter int
	params      []interface{}
}

func NewQuery() *query {
	return &query{
		numOfFilter: 0,
	}
}

func (q *query) BaseQueryCountSelectAllUser() *query {
	q.query = `
	SELECT
		COUNT(u.id)
	FROM
		users u
	`
	return q
}

func (q *query) BaseQuerySelectAllUser() *query {
	q.query = `
	SELECT
		u.id,
		u.name,
		u.email,
		u.phone_number,
		u.coin,
		ST_X(u.coordinate),
		ST_Y(u.coordinate),
		u.is_email_verified,
		u.email_verified_at,
		u.is_driver,
		u.created_at,
		u.updated_at
	FROM users u
	`
	return q
}

func (q *query) AddWhereClause() *query {
	q.whereExist = true
	q.query += ` WHERE `
	return q
}

func (q *query) AddFilter(column string, value interface{}) *query {
	if !q.whereExist {
		q.AddWhereClause()
	}
	if q.numOfFilter != 0 {
		q.AddAndClause()
	}
	q.numOfFilter++
	q.query += fmt.Sprintf(`%s = ?`, column)
	q.params = append(q.params, value)
	return q
}

func (q *query) AddAndClause() *query {
	q.query += ` AND `
	return q
}

func (q *query) AddLimit(limit int64) *query {
	q.query += fmt.Sprintf(` LIMIT %d`, limit)
	return q
}

func (q *query) AddOffset(offset int64) *query {
	q.query += fmt.Sprintf(` OFFSET %d`, offset)
	return q
}

func (q *query) AddOrderBy(column string, order string) *query {
	q.query += fmt.Sprintf(` ORDER BY %s %s`, column, order)
	return q
}

func (q *query) GetQuery() string {
	return q.query
}

func (q *query) GetParams() []interface{} {
	return q.params
}
