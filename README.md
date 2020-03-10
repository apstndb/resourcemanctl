List projects in organization with billing account and folder path

## Example output

```
$ go run ./ --organization=${ORGANIZATION_ID} --output=result.csv
$ head result.csv
project_id,billing_account_id,display_name_path
test-project,billingAccounts/123456-7890AB-CDEFGH,/
test-project-child,billingAccounts/123456-7890AB-CDEFGH,/child
test-project-grandchildren,billingAccounts/123456-7890AB-CDEFGH,/children/grandchildren
```
