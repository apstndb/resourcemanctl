List projects in organization with billing account and folder path.
It takes long time because projects.getBillingInfo is called for each projects sequentially.

## Example output

```
$ go run ./ --organization=${ORGANIZATION_ID} --output=result.csv
$ head result.csv
project_id,billing_account_id,display_name_path
test-project,billingAccounts/123456-7890AB-CDEFGH,/
test-project-child,billingAccounts/123456-7890AB-CDEFGH,/children
test-project-grandchildren,billingAccounts/123456-7890AB-CDEFGH,/children/grandchildren
```
