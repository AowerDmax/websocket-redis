package KnowledgeDatabase

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"websocket-redis/config"

	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"github.com/xuri/excelize/v2"
)

type QA struct {
	ID string `json:"id"`
	Q  string `json:"q"`
	A  string `json:"a"`
}

func ProcessExcel() {
	cfg := config.LoadConfig()
	meilisearchHost := fmt.Sprintf("http://%s:%d", cfg.MeiliSearchHost, cfg.MeiliSearchPort)

	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host: meilisearchHost,
	})

	index := client.Index("qa_pairs")
	_, err := index.FetchPrimaryKey()
	if err != nil {
		_, createErr := client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        "qa_pairs",
			PrimaryKey: "id",
		})
		if createErr != nil {
			log.Fatalf("Error creating index and setting primary key: %v", createErr)
		}
	}

	folderPath := "./data"

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error walking through files: %v", err)
			return err
		}

		if filepath.Ext(path) == ".xlsx" {
			readExcelFile(path, client)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking through directory: %v", err)
	}
}

func readExcelFile(filePath string, client *meilisearch.Client) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatalf("Error opening Excel file: %v", err)
	}

	sheets := f.GetSheetList()

	for _, sheet := range sheets {
		rows, err := f.GetRows(sheet)
		if err != nil {
			log.Fatalf("Error reading rows from sheet: %v", err)
		}

		var qaPairs []QA
		for _, row := range rows[1:] {
			if len(row) >= 2 {
				qa := QA{
					ID: uuid.New().String(),
					Q:  row[0],
					A:  row[1],
				}
				qaPairs = append(qaPairs, qa)
			}
		}

		if len(qaPairs) > 0 {
			index := client.Index("qa_pairs")
			_, err = index.AddDocuments(qaPairs, "id")
			if err != nil {
				log.Fatalf("Error storing data to MeiliSearch: %v", err)
			}
		}
	}
}
