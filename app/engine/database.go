package engine

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/zhenghaoz/gorse/core"
	"io"
)

type Database struct {
	connection *sql.DB
}

func NewDatabaseConnection(databaseDriver, dataSource string) (db Database, err error) {
	db.connection, err = sql.Open(databaseDriver, dataSource)
	return
}

func (db *Database) Close() error {
	return db.connection.Close()
}

// Initialize the SQL database for gorse. Three tables will be created:
// 1. ratings: all ratings given by users to items;
// 2. items: all items will be recommended to users;
// 3. recommends: recommended items for each user.
func (db *Database) Init() error {
	// Create ratings table
	_, err := db.connection.Exec(`CREATE TABLE ratings (
			user_id int NOT NULL,
			item_id int NOT NULL,
			rating int NOT NULL,
			UNIQUE KEY unique_index (user_id,item_id)
		)`)
	if err != nil {
		return err
	}
	// Create recommends table
	_, err = db.connection.Exec(`CREATE TABLE recommends (
			user_id int NOT NULL,
			item_id int NOT NULL,
			ranking double NOT NULL,
			UNIQUE KEY unique_index (user_id,item_id)
		)`)
	if err != nil {
		return err
	}
	// Create items table
	_, err = db.connection.Exec(`CREATE TABLE items (
			item_id int NOT NULL,
			UNIQUE KEY unique_index item_id
		)`)
	return err
}

func (db *Database) LoadItemsFromCSV(fileName string, sep string, header bool) error {
	return nil
}

func (db *Database) LoadRatingsFromCSV(fileName string, sep string, header bool) error {
	return nil
}

func (db *Database) GetMeta(name string) (count int, err error) {
	// Query SQL
	rows, err := db.connection.Query("SELECT value FROM status WHERE name = '?'", name)
	if err != nil {
		return
	}
	// Retrieve result
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return
		}
		return
	}
	panic("Get meta data failed")
}

func (db *Database) SetMeta(name string, val int) error {
	panic("Not implemented")
}

// CurrentRatings gets the number of ratings at current.
func (db *Database) CurrentRatings() (count int, err error) {
	rows, err := db.connection.Query("SELECT COUNT(*) FROM ratings")
	if err != nil {
		return
	}
	// Retrieve result
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return
		}
		return
	}
	panic("SELECT COUNT(*) FROM ratings failed")
}

// LastRatings gets the number of ratings at the time of last update.
func (db *Database) LastRatings() (count int, err error) {
	return db.GetMeta("last_count")
}

func (db *Database) Version() (version int, err error) {
	return db.GetMeta("version")
}

func (db *Database) LoadData() (*core.DataSet, error) {
	return core.LoadDataFromSQL(db.connection, "ratings", "user_id", "item_id", "rating")
}

// GetRecommends gets the top list for a user from the database.
func (db *Database) GetRecommends(userId int) ([]int, error) {
	// Query SQL
	rows, err := db.connection.Query("SELECT item_id FROM recommends WHERE user_id=? ORDER BY rating DESC;", userId)
	if err != nil {
		return nil, err
	}
	// Retrieve result
	res := make([]int, 0)
	for rows.Next() {
		var itemId int
		err = rows.Scan(&itemId)
		if err != nil {
			return nil, err
		}
		res = append(res, itemId)
	}
	return res, nil
}

func (db *Database) GetRandom() ([]int, error) {
	panic("Not implemented")
}

func (db *Database) GetPopular() ([]int, error) {
	panic("Not implemented")
}

func (db *Database) GetList() ([]int, error) {
	panic("Not implemented")
}

// UpdateRecommends puts a top list into the database.
func (db *Database) UpdateRecommends(userId int, items []int) error {
	buf := bytes.NewBuffer(nil)
	for i, itemId := range items {
		buf.WriteString(fmt.Sprintf("%d\t%d\t%d\n", userId, itemId, i))
	}
	mysql.RegisterReaderHandler("update_recommends", func() io.Reader {
		return bytes.NewReader(buf.Bytes())
	})
	_, err := db.connection.Exec("LOAD DATA LOCAL INFILE 'Reader::update_recommends' INTO TABLE recommends")
	return err
}

// PutRating puts a rating into the database.
func (db *Database) PutRating(userId, itemId int, rating float64) error {
	// Prepare SQL
	statement, err := db.connection.Prepare("INSERT INTO ratings VALUES(?, ?, ?) ON DUPLICATE KEY UPDATE rating=VALUES(rating)")
	if err != nil {
		return err
	}
	// Execute SQL
	_, err = statement.Exec(userId, itemId, rating)
	if err != nil {
		return err
	}
	return nil
}