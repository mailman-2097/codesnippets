package main

import (
	"os"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/organizations"

	"fmt"
)

// OU : Struct
type OU struct {
	OuId     string
	OuName   string
	ParentId string
}

func importOUs(orgRootId string) {
	var ous map[string]OU
	ous = listOUs(orgRootId)
	CmdExec("bash", "-c", "> imports_organizations.tf")
	for k := range ous {
		SaveTFOUResource(ous[k])
		fmt.Printf("terraform import aws_organizations_organizational_unit %s\n", ous[k].OuId)
		CmdExec("terraform", "import", fmt.Sprintf("aws_organizations_organizational_unit.%s", ous[k].OuId), fmt.Sprintf("%s", ous[k].OuId))
		importSCPattachments(ous[k].OuId)
	}
}
func listOUs(rootId string) map[string]OU {
	svc := organizations.New(session.New())
	input := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(rootId),
	}

	result, err := svc.ListOrganizationalUnitsForParent(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case organizations.ErrCodeAccessDeniedException:
				fmt.Println(organizations.ErrCodeAccessDeniedException, aerr.Error())
			case organizations.ErrCodeAWSOrganizationsNotInUseException:
				fmt.Println(organizations.ErrCodeAWSOrganizationsNotInUseException, aerr.Error())
			case organizations.ErrCodeInvalidInputException:
				fmt.Println(organizations.ErrCodeInvalidInputException, aerr.Error())
			case organizations.ErrCodeParentNotFoundException:
				fmt.Println(organizations.ErrCodeParentNotFoundException, aerr.Error())
			case organizations.ErrCodeServiceException:
				fmt.Println(organizations.ErrCodeServiceException, aerr.Error())
			case organizations.ErrCodeTooManyRequestsException:
				fmt.Println(organizations.ErrCodeTooManyRequestsException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}

	fmt.Println(result)
	results := make(map[string]OU)
	for _, ou := range result.OrganizationalUnits {
		results[aws.StringValue(ou.Id)] = OU{OuId: aws.StringValue(ou.Id), OuName: aws.StringValue(ou.Name), ParentId: rootId}
	}
	fmt.Println("OU(s) identified:", len(results))
	return results
}

// SaveTFOUResource function
func SaveTFOUResource(ou OU) {
	const tfScppolicyTemplate = `
	resource "aws_organizations_organizational_unit" {{.OuId}} {
		name      = "{{.OuName}}"
		parent_id = "{{.ParentId}}"
	}
	`
	fmt.Printf("OU => %+v \n", ou)
	t := template.Must(template.New("aws_organizations_organizational_unit").Parse(tfScppolicyTemplate))
	// err := t.Execute(os.Stdout, ou)
	// errCheck("SaveTFOUResource Templating Error :", err)
	f, err := os.OpenFile(ouFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	ioErrCheck("SaveTFOUResource Error File Open:", err)
	err = t.Execute(f, ou)
	ioErrCheck("SaveTFOUResource Error Persist IO:", err)
	f.Close()
}
