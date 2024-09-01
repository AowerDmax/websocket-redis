package KnowledgeDatabase

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
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
	log.Printf("RAG Connect Processing...")
	client, connected := tryConnectMeilisearch(cfg)

	folderPath := "./data"
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error walking through files: %v", err)
			return err
		}
		if filepath.Ext(path) == ".xlsx" {
			readExcelFile(path, client, connected)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error walking through directory: %v", err)
	}
}

func tryConnectMeilisearch(cfg *config.Config) (*meilisearch.Client, bool) {
	meilisearchHost := fmt.Sprintf("http://%s:%d", cfg.MeiliSearchHost, cfg.MeiliSearchPort)
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host: meilisearchHost,
	})

	for attempts := 1; attempts <= 3; attempts++ {
		_, err := client.Health()
		if err == nil {
			index := client.Index("qa_pairs")
			_, err := index.FetchInfo()
			if err != nil {
				_, createErr := client.CreateIndex(&meilisearch.IndexConfig{
					Uid:        "qa_pairs",
					PrimaryKey: "id",
				})
				if createErr != nil {
					log.Printf("Error creating index: %v", createErr)
					return client, false
				}
			}
			return client, true
		}
		log.Printf("Attempt %d: Failed to connect to Meilisearch. Retrying in 5 seconds...", attempts)
		time.Sleep(5 * time.Second)
	}

	log.Println("Failed to connect to Meilisearch after 3 attempts. Proceeding without Meilisearch.")
	return client, false
}

func readExcelFile(filePath string, client *meilisearch.Client, meilisearchConnected bool) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Printf("Error opening Excel file %s: %v", filePath, err)
		return
	}
	defer f.Close()

	sheets := f.GetSheetList()
	for _, sheet := range sheets {
		rows, err := f.GetRows(sheet)
		if err != nil {
			log.Printf("Error reading rows from sheet %s: %v", sheet, err)
			continue
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
		if len(qaPairs) > 0 && meilisearchConnected {
			index := client.Index("qa_pairs")
			_, err = index.AddDocuments(qaPairs)
			if err != nil {
				log.Printf("Error storing data to MeiliSearch: %v", err)
			} else {
				log.Printf("Successfully added %d QA pairs to Meilisearch from sheet %s", len(qaPairs), sheet)
			}
		} else if len(qaPairs) > 0 {
			log.Printf("Processed %d QA pairs from sheet %s (not stored in Meilisearch)", len(qaPairs), sheet)
		}
	}
}
