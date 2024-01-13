# incident.io cli

```bash
go build -o inc .
export INC_API_KEY=inc_foobarbaz

# get all incidents
inc incident get
# get an incident by reference number, e.g. INC-123
inc incident get --reference 123
# get an incident by id
inc incident get --id 01HE6...

# set custom field Oncall Rotation to Serving Infra Default
inc incident edit --reference 123 --field "Oncall Rotation=Serving Infra Default"

# remove an existing custom field by passing 'NAME=` with no value
inc incident edit --reference 123  --field "Oncall Rotation="

# set custom field foo to 'bar=baz', after the first equal sign the text is used as value verbatim
inc incident edit --id 01HE6...   --field "foo=bar=baz"

# list all catalog types
inc catalog types get 
# list catalog types where name == Roles
inc catalog types get --name Roles
# show catalog type with id == 01He6...
inc catalog types get --id 01He6...

# list all catalog entries across all types
inc catalog entries get 
# find a catalog entry by name and type name, enumerating types for first match and returning 1 entry match for that type.
inc catalog entries get --name NAME --type-name TYPE_NAME
# get a catalog entry by id
inc catalog entries get --id 01He6...
# find a catalog entry by name, returning all matches across all types.
inc catalog entries get --name NAME
```
