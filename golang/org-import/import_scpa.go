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

// Required for Text\template : Struct
type TargetSCPAttachment struct {
	PolicyId     string
	TargetId     string
	AttachmentId string
}

func importSCPattachments(targetId string) {
	attachments := listPoliciesForTarget(targetId)
	CmdExec("bash", "-c", "> imports_attachments.tf")
	for _, at := range attachments {
		SaveTFSCPAttachmentResource(at)
		fmt.Printf("terraform import aws_organizations_policy_attachment %s\n", at.AttachmentId)
		CmdExec("terraform", "import", fmt.Sprintf("aws_organizations_policy_attachment.%s", at.AttachmentId), fmt.Sprintf("%s:%s", at.TargetId, at.PolicyId))
	}
}
func listPoliciesForTarget(targetId string) []TargetSCPAttachment {
	svc := organizations.New(session.New())
	input := &organizations.ListPoliciesForTargetInput{
		Filter:   aws.String("SERVICE_CONTROL_POLICY"),
		TargetId: aws.String(targetId),
	}

	result, err := svc.ListPoliciesForTarget(input)
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
			case organizations.ErrCodeTargetNotFoundException:
				fmt.Println(organizations.ErrCodeTargetNotFoundException, aerr.Error())
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

	count := len(result.Policies)
	results := make([]TargetSCPAttachment, count)
	for k, p := range result.Policies {
		results[k] = TargetSCPAttachment{AttachmentId: fmt.Sprintf("SCPATTCH%d-%s", k, targetId), PolicyId: aws.StringValue(p.Id), TargetId: targetId}
	}
	fmt.Println("Attachment(s) identified:", count)
	return results
}

// SaveTFSCPAttachmentResource function
func SaveTFSCPAttachmentResource(at TargetSCPAttachment) {
	const tfScpAttachmentTemplate = `
	resource "aws_organizations_policy_attachment" {{.AttachmentId}} {
		policy_id = "{{.PolicyId}}"
		target_id = "{{.TargetId}}"
	}
	`
	t := template.Must(template.New("aws_organizations_policy_attachment").Parse(tfScpAttachmentTemplate))
	// err := t.Execute(os.Stdout, at)
	// errCheck("SaveTFSCPAttachmentResource Templating Error :", err)
	f, err := os.OpenFile(attachmentsFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	ioErrCheck("SaveTFSCPAttachmentResource Error File Open:", err)
	err = t.Execute(f, at)
	ioErrCheck("SaveTFSCPAttachmentResource Error Persist IO:", err)
	f.Close()
}
