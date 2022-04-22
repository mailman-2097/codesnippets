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

// ORG : Struct
type ORG struct {
	OrgId          string
	OrgName        string
	OrgPolicyTypes []OrgPolicyTypes
}

type OrgPolicyTypes struct {
	Status string
	Type   string
}

func importRoot() string {
	var root map[string]ORG
	root = listRoot()
	CmdExec("bash", "-c", "> imports_roots.tf")
	var orgRootId string
	for k := range root {
		SaveTFRootResource(root[k])
		fmt.Printf("terraform import aws_organizations_organization %s\n", root[k].OrgName)
		CmdExec("terraform", "import", fmt.Sprintf("aws_organizations_organization.%s", root[k].OrgName), fmt.Sprintf("%s", root[k].OrgId))
		orgRootId = root[k].OrgId
	}
	importSCPattachments(orgRootId)
	return orgRootId
}
func listRoot() map[string]ORG {
	svc := organizations.New(session.New())
	input := &organizations.ListRootsInput{}

	result, err := svc.ListRoots(input)
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

	results := make(map[string]ORG)
	for _, root := range result.Roots {
		pTypes := make([]OrgPolicyTypes, len(root.PolicyTypes))
		for k, pType := range root.PolicyTypes {
			pTypes[k] = OrgPolicyTypes{Status: aws.StringValue(pType.Status), Type: aws.StringValue(pType.Type)}
		}
		results[aws.StringValue(root.Id)] = ORG{OrgId: aws.StringValue(root.Id), OrgName: aws.StringValue(root.Name), OrgPolicyTypes: pTypes}
		break // first root only
	}
	fmt.Println("Root(s) identified:", len(results))
	return results
}

// SaveTFRootResource function
func SaveTFRootResource(root ORG) {
	const tfScppolicyTemplate = `
	resource "aws_organizations_organization" {{.OrgName}} {
		enabled_policy_types = [
			{{range .OrgPolicyTypes}} "{{.Type}}", {{end}}
		]

		feature_set = "ALL"
	}
	`
	t := template.Must(template.New("aws_organizations_organization").Parse(tfScppolicyTemplate))
	// err := t.Execute(os.Stdout, root)
	// errCheck("Org Templating Error :", err)
	f, err := os.OpenFile(rootFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	ioErrCheck("SaveTFRootResource Error File Open:", err)
	err = t.Execute(f, root)
	ioErrCheck("SaveTFRootResource Error Persist IO:", err)
	f.Close()
}
