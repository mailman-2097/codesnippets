#!/bin/bash

set -eu
> imports_scps.tf
> imports_accts.tf
> imports_ous.tf
> imports_org.tf
> imports_scpas.tf

function import_scp_attachment () {
  local target_id=$1
  local j=0
  local scp_attachment_name=""
  target_scp_attachment_ids=($(aws organizations list-policies-for-target --filter SERVICE_CONTROL_POLICY --query 'Policies[*].[Id]' --target-id $target_id --output=text))
  echo "Attachments for target $target_id" "${target_scp_attachment_ids[@]}"
  for ((; j < "${#target_scp_attachment_ids[@]}"; j++)); do
    scp_attachment_name="SCPATT${j}-${target_id}"
    printf 'resource "aws_organizations_policy_attachment" %s {}\n' "${scp_attachment_name}" >> imports_scpas.tf
    printf 'terraform import aws_organizations_policy_attachment.%s %s\n' "${scp_attachment_name}" "${target_scp_attachment_ids[j]}"
    terraform import aws_organizations_policy_attachment."${scp_attachment_name}" "${target_id}:${target_scp_attachment_ids[j]}"
  done
}

# aws organizations list-policies --filter SERVICE_CONTROL_POLICY --query 'Policies[*].[Name, Id]' --output=text
scp_ids=($(aws organizations list-policies --filter SERVICE_CONTROL_POLICY --query 'Policies[*].[Id]' --output=text))
scp_names=($(aws organizations list-policies --filter SERVICE_CONTROL_POLICY --query 'Policies[*].[Name]' --output=text))
echo "Count of SCPs to be imported" "${#scp_ids[@]}"
for ((i = 0; i < "${#scp_ids[@]}"; i++)); do
  printf 'resource "aws_organizations_policy" %s {}\n' "${scp_names[i]}" >> imports_scps.tf
  printf 'terraform import aws_organizations_policy.%s %s\n' "${scp_names[i]}" "${scp_ids[i]}"
  terraform import aws_organizations_policy."${scp_names[i]}" "${scp_ids[i]}"
done
echo "---------------------------------------------"

# # aws organizations list-accounts --query 'Accounts[*].[Id, Arn]' --output=text
account_ids=($(aws organizations list-accounts --query 'Accounts[*].[Id]' --output=text))
echo "Count of Accounts to be imported" "${#account_ids[@]}"
for ((i = 0; i < "${#account_ids[@]}"; i++)); do
  printf 'resource "aws_organizations_account" %s {}\n' "A${account_ids[i]}" >> imports_accts.tf
  printf 'terraform import aws_organizations_account.%s %s\n' "A${account_ids[i]}" "${account_ids[i]}"
  terraform import aws_organizations_account."A${account_ids[i]}" "${account_ids[i]}"
  import_scp_attachment "${account_ids[i]}"
done
echo "---------------------------------------------"

# # aws organizations list-roots --query 'Roots[*].[Id]' --output=text
root_id=$(aws organizations list-roots --query 'Roots[*].[Id]' --output=text)
root_name=$(aws organizations list-roots --query 'Roots[*].[Name]' --output=text)
printf 'resource "aws_organizations_organization" %s {}\n' "$root_name" >> imports_org.tf
printf 'terraform import aws_organizations_organization.%s %s\n' "$root_name" "$root_id"
import_scp_attachment "$root_id"
terraform import aws_organizations_organization."$root_name" "$root_id"

# # aws organizations list-organizational-units-for-parent --parent-id r-0n1u --output=text
org_ou_ids=($(aws organizations list-organizational-units-for-parent  --query 'OrganizationalUnits[*].[Id]' --parent-id "$root_id" --output=text))
echo "Count of OUs to be imported" "${#org_ou_ids[@]}"
for ((i = 0; i < "${#org_ou_ids[@]}"; i++)); do
  printf 'resource "aws_organizations_organizational_unit" %s {}\n' "${org_ou_ids[i]}" >> imports_ous.tf
  printf 'terraform import aws_organizations_organizational_unit.%s %s\n' "${org_ou_ids[i]}" "${org_ou_ids[i]}"
  terraform import aws_organizations_organizational_unit."${org_ou_ids[i]}" "${org_ou_ids[i]}"
  import_scp_attachment "${org_ou_ids[i]}"
done
echo "---------------------------------------------"
