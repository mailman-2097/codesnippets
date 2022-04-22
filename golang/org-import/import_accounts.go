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

// Account : Struct
type Account struct {
	Name  string
	Email string
	Id    string
}

func importAccounts() {
	var accounts map[string]Account
	accounts = listAccounts()
	CmdExec("bash", "-c", "> imports_accounts.tf")
	for k := range accounts {
		SaveTFAccountResource(accounts[k])
		fmt.Printf("terraform import aws_organizations_account %s\n", accounts[k].Name)
		CmdExec("terraform", "import", fmt.Sprintf("aws_organizations_account.A%s", accounts[k].Id), fmt.Sprintf("%s", accounts[k].Id))
		importSCPattachments(accounts[k].Id)
	}
}
func listAccounts() map[string]Account {
	svc := organizations.New(session.New())
	input := &organizations.ListAccountsInput{}

	result, err := svc.ListAccounts(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case organizations.ErrCodeAccessDeniedException:
				fmt.Println(organizations.ErrCodeAccessDeniedException, aerr.Error())
			case organizations.ErrCodeAWSOrganizationsNotInUseException:
				fmt.Println(organizations.ErrCodeAWSOrganizationsNotInUseException, aerr.Error())
			case organizations.ErrCodeInvalidInputException:
				fmt.Println(organizations.ErrCodeInvalidInputException, aerr.Error())
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
	results := make(map[string]Account)
	for _, ac := range result.Accounts {
		results[aws.StringValue(ac.Id)] = Account{Id: aws.StringValue(ac.Id), Name: aws.StringValue(ac.Name), Email: aws.StringValue(ac.Email)}
	}
	fmt.Println("Accounts(s) identified:", len(results))
	fmt.Printf("Accounts => %+v \n", results)
	return results
}

// SaveTFAccountResource function
func SaveTFAccountResource(ac Account) {
	const tfScppolicyTemplate = `
	resource "aws_organizations_account" A{{.Id}} {
		name  = "{{.Name}}"
		email = "{{.Email}}"
	}
	`
	fmt.Printf("AC => %+v \n", ac)
	t := template.Must(template.New("aws_organizations_account").Parse(tfScppolicyTemplate))
	// err := t.Execute(os.Stdout, ac)
	// errCheck("SaveTFAccountResource Templating Error :", err)
	f, err := os.OpenFile(acFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	ioErrCheck("SaveTFAccountResource Error File Open:", err)
	err = t.Execute(f, ac)
	ioErrCheck("SaveTFAccountResource Error Persist IO:", err)
	f.Close()
}
