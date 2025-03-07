package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type PlanStorageAzure struct {
	Container string
	Client    *azblob.Client
	//	container  *storage.containerHandle
	Context context.Context
}

func NewAzurePlanStorage(storageAccountName string, container string) (*PlanStorageAzure, error) {
	url := fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccountName)
	ctx := context.Background()

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	client, err := azblob.NewClient(url, credential, nil)
	if err != nil {
		return nil, err
	}

	return &PlanStorageAzure{
		Container: container,
		Client:    client,
		Context:   ctx,
	}, nil
}

func (psa *PlanStorageAzure) PlanExists(artifactName string, storedPlanFilePath string) (bool, error) {
	_, err := psa.Client.ServiceClient().NewContainerClient(psa.Container).NewBlobClient(storedPlanFilePath).GetProperties(psa.Context, nil)
	if err != nil {
		return false, fmt.Errorf("unable to get object attributes: %v", err)
	}

	return true, nil
}

func (psa *PlanStorageAzure) StorePlanFile(fileContents []byte, artifactName string, fileName string) error {
	if _, err := psa.Client.UploadBuffer(psa.Context, psa.Container, fileName, fileContents, nil); err != nil {
		log.Printf("Failed to write file to container: %v", err)
		return err
	}
	return nil
}

func (psa *PlanStorageAzure) RetrievePlan(localPlanFilePath string, artifactName string, storedPlanFilePath string) (*string, error) {
	file, err := os.Create(localPlanFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %v", err)
	}
	defer file.Close()

	_, err = psa.Client.DownloadFile(psa.Context, psa.Container, storedPlanFilePath, file, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to write data to file: %v", err)
	}
	fileName, err := filepath.Abs(file.Name())
	if err != nil {
		return nil, fmt.Errorf("unable to get absolute path for file: %v", err)
	}
	return &fileName, nil
}

func (psa *PlanStorageAzure) DeleteStoredPlan(artifactName string, storedPlanFilePath string) error {
	_, err := psa.Client.ServiceClient().NewContainerClient(psa.Container).NewBlobClient(storedPlanFilePath).Delete(psa.Context, nil)

	if err != nil {
		return fmt.Errorf("unable to delete file '%v' from container: %v", storedPlanFilePath, err)
	}
	return nil
}
