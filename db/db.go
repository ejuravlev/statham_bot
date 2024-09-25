package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Db struct {
	connection *sql.DB
}

type Quote struct {
	Id   int
	Text string
}

func NewConnection(connectionString string) (*Db, error) {
	connection, err := sql.Open("postgres", connectionString)
	if err != nil {
		return &Db{}, err
	}

	return &Db{connection: connection}, nil
}

func (d *Db) GetRandomStathamQuote() Quote {
	q := Quote{}
	row := d.connection.QueryRow("SELECT id, quote FROM quotes ORDER BY random() LIMIT 1;")
	err := row.Scan(&q.Id, &q.Text)
	if err != nil {
		panic(err)
	}
	return q
}

func (d *Db) SetNewQuote(text string) error {
	_, err := d.connection.Exec("INSERT INTO quotes (text) values ('$1');", text)
	if err != nil {
		return fmt.Errorf("не удалось произвести вставку цитаты в цитатник: %s", err.Error())
	}
	return nil
}

func (d *Db) GetSubscibers() ([]int64, error) {
	ids := make([]int64, 0)
	rows, err := d.connection.Query("SELECT id FROM chats WHERE subscribed IS TRUE;")
	if err != nil {
		return ids, fmt.Errorf("не удалось произвести поиск подписчиков: %s", err.Error())
	}

	var id int64
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return ids, fmt.Errorf("не удалось считать результат поиска подпичиков: %s", err.Error())
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (d *Db) SubscribeChat(id int64) error {
	_, err := d.connection.Exec("INSERT INTO chats (id, subscribed) values ($1, $2) ON CONFLICT (id) DO UPDATE SET subscribed = $2;", id, true)
	if err != nil {
		return fmt.Errorf("не удалось подписать пользователя: %s", err.Error())
	}
	return nil
}

func (d *Db) UnsubscribeChat(id int64) error {
	_, err := d.connection.Exec("INSERT INTO chats (id, subscribed) values ($1, $2) ON CONFLICT (id) DO UPDATE SET subscribed = $2;", id, false)
	if err != nil {
		return fmt.Errorf("не удалось отписать пользователя: %s", err.Error())
	}
	return nil
}
