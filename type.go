package main

import (
	"log"
	"time"
)

type Item struct {
	ItemID        string    `json:"id"`
	Title         string    `json:"title"`
	CategoryID    string    `json:"category_id"`
	Price         int       `json:"price"`
	StartTime     time.Time `json:"start_time"`
	StopTime      time.Time `json:"stop_time"`
	Children      []Child   `json:"children"`
	cacheDatabase *Database
}

func (i *Item) SetCache() *Item {
	database := Database{}
	i.cacheDatabase = database.GetClient()
	sqlStatement := `INSERT INTO item (id, title, category_id, price,start_time,stop_time)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id`
	id := ""
	err := i.cacheDatabase.Client.QueryRow(sqlStatement,
		i.ItemID,
		i.Title,
		i.CategoryID,
		i.Price,
		i.StartTime,
		i.StartTime).Scan(&id)
	if err != nil {
		log.Printf("%v", err)
		return nil
	}
	log.Printf("New ITEM ID is: %v", id)

	if id != "" {
		sqlStatement = `INSERT INTO child (id, parent_id, stop_time)
VALUES ($1, $2, $3)
RETURNING id`
		for _, child := range i.Children {
			id := ""
			err := i.cacheDatabase.Client.QueryRow(sqlStatement,
				child.ItemID,
				i.ItemID,
				child.StopTime).Scan(&id)
			if err != nil {
				log.Printf("%v", err)
				return nil
			}
			log.Printf("New CHILD ID is: %v", id)
		}
	}
	return i
}

func (i *Item) GetCache() (error, *Item) {
	database := Database{}
	i.cacheDatabase = database.GetClient()
	sqlStatement := `SELECT * FROM item WHERE id = $1`

	err := i.cacheDatabase.Client.QueryRow(sqlStatement, i.ItemID).
		Scan(&i.ItemID,
			&i.Title,
			&i.CategoryID,
			&i.Price,
			&i.StartTime,
			&i.StopTime)
	if err != nil {
		log.Printf("%v", err)
		return err, nil
	}

	sqlStatement = `SELECT c.id, c.stop_time FROM item i, child c WHERE i.id = $1 AND c.parent_id = $1`

	rows, err := i.cacheDatabase.Client.Query(sqlStatement, i.ItemID)

	if err != nil {
		log.Printf("Query: %v", err)
		return err, nil
	}

	var children []Child

	for rows.Next() {

		var child Child
		err = rows.Scan(&child.ItemID, &child.StopTime)

		if err != nil {

			log.Printf("Scan: %v", err)
			return err, nil
		}

		children = append(children, child)

	}
	err = rows.Err()
	if err != nil {
		log.Printf("Err: %v", err)
		return err, nil
	}
	i.Children = children

	return nil, i

}

type Child struct {
	ItemID   string    `json:"id,itemized"`
	StopTime time.Time `json:"stop_time"`
}

type Health struct {
	Date                    time.Time     `json:"date"`
	AvgResponseTime         float64       `json:"avg_response_time"`
	TotalRequest            float64       `json:"total_requests"`
	AvgResponseTimeApiCalls float64       `json:"avg_response_time_api_calls"`
	TotalCountApiCall       float64       `json:"total_count_api_calls"`
	InfoRequests            []InfoRequest `json:"info_requests"`
}

type InfoRequest struct {
	StatusCode int `json:"status_code"`
	Count      int `json:"count"`
}
