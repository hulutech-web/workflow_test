package facades

import (
	"goravel/packages/goravel-workflow"
	"goravel/packages/goravel-workflow/contracts"
	"log"
)

func Workflow() contracts.Workflow {
	instance, err := workflow.App.Make(workflow.Binding)
	if err != nil {
		log.Println(err)
		return nil
	}

	return instance.(contracts.Workflow)
}
