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

// SCPPolicy : Struct
type SCPPolicy struct {
	PolicyId     string
	PolicyName   string
	PolicyString string
}

func importScps() {
	scpList := listScpPolicies()
	for k := range scpList {
		p := scpList[k]
		p.PolicyString = fetchScpDetails(k)
		scpList[k] = p
	}
	// for k := range scpList {
	// 	fmt.Printf("Policy Id = %s => %+v \n", k, scpList[k])
	// }
	CmdExec("bash", "-c", "> imports_policies.tf")
	for k := range scpList {
		SaveTFSCPResource(scpList[k])
		fmt.Printf("terraform import aws_organizations_policy.%s %s\n", scpList[k].PolicyName, scpList[k].PolicyId)
		CmdExec("terraform", "import", fmt.Sprintf("aws_organizations_policy.%s", scpList[k].PolicyName), scpList[k].PolicyId)
	}
	CmdExec("terraform", "fmt")
}

func listScpPolicies() map[string]SCPPolicy {

	svc := organizations.New(session.New())
	input := &organizations.ListPoliciesInput{
		Filter: aws.String("SERVICE_CONTROL_POLICY"),
	}

	result, err := svc.ListPolicies(input)
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
			case organizations.ErrCodeUnsupportedAPIEndpointException:
				fmt.Println(organizations.ErrCodeUnsupportedAPIEndpointException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}
	results := make(map[string]SCPPolicy)
	for _, p := range result.Policies {
		results[aws.StringValue(p.Id)] = SCPPolicy{PolicyId: aws.StringValue(p.Id), PolicyName: aws.StringValue(p.Name), PolicyString: ""}
	}
	fmt.Println("Scps identified:", len(results))
	return results
}

// fetchScpDetails : fetch policy string
func fetchScpDetails(policyId string) string {
	svc := organizations.New(session.New())
	input := &organizations.DescribePolicyInput{
		PolicyId: aws.String(policyId),
	}

	result, err := svc.DescribePolicy(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case organizations.ErrCodeAccessDeniedException:
				fmt.Println(organizations.ErrCodeAccessDeniedException, aerr.Error())
			case organizations.ErrCodeAWSOrganizationsNotInUseException:
				fmt.Println(organizations.ErrCodeAWSOrganizationsNotInUseException, aerr.Error())
			case organizations.ErrCodeInvalidInputException:
				fmt.Println(organizations.ErrCodeInvalidInputException, aerr.Error())
			case organizations.ErrCodePolicyNotFoundException:
				fmt.Println(organizations.ErrCodePolicyNotFoundException, aerr.Error())
			case organizations.ErrCodeServiceException:
				fmt.Println(organizations.ErrCodeServiceException, aerr.Error())
			case organizations.ErrCodeTooManyRequestsException:
				fmt.Println(organizations.ErrCodeTooManyRequestsException, aerr.Error())
			case organizations.ErrCodeUnsupportedAPIEndpointException:
				fmt.Println(organizations.ErrCodeUnsupportedAPIEndpointException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}
	return string(aws.StringValue(result.Policy.Content))
}

// SaveTFSCPResource function
func SaveTFSCPResource(scpPolicy SCPPolicy) {
	const tfScppolicyTemplate = `
	resource "aws_organizations_policy" {{.PolicyName}} {
		name = "{{.PolicyName}}"
		content = <<CONTENT
		{{.PolicyString}}
		CONTENT
	}`
	// fmt.Printf("Policy => %+v \n", scpPolicy)
	t := template.Must(template.New("aws_organizations_policy").Parse(tfScppolicyTemplate))
	// err := t.Execute(os.Stdout, scpPolicy)
	// errCheck("SCP Policy Templating Error :", err)
	f, err := os.OpenFile(scpFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	ioErrCheck("SaveTFSCPResource Error File Open:", err)
	err = t.Execute(f, scpPolicy)
	ioErrCheck("SaveTFSCPResource Error Persist IO:", err)
	f.Close()
}
