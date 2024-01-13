package main

import (
	"context"
	"fmt"

	"github.com/alexeldeib/incli/client"
	kitlog "github.com/go-kit/log"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sanity-io/litter"
)

func FindCustomFieldByName(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, targetName string) (*client.CustomFieldV2, error) {
	customFields, err := ListAllCustomFields(ctx, logger, cl)
	if err != nil {
		return nil, errors.Wrap(err, "listing custom fields types")
	}

	for _, v := range customFields {
		if v.Name == targetName {
			return &v, nil
		}
	}

	return nil, errors.Errorf("catalog type %q not found", targetName)
}

func FindCatalogTypeByID(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, targetID string) (*client.CatalogTypeV2, error) {
	catalogTypes, err := ListAllCatalogTypes(ctx, logger, cl)
	if err != nil {
		return nil, fmt.Errorf("finding catalog type by name: %s", err)
	}

	for _, v := range catalogTypes {
		if v.Id == targetID {
			return &v, nil
		}
	}

	return nil, errors.Errorf("catalog type %q not found", targetID)
}

func FindCatalogTypeByName(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, targetName string) (*client.CatalogTypeV2, error) {
	catalogTypes, err := ListAllCatalogTypes(ctx, logger, cl)
	if err != nil {
		return nil, fmt.Errorf("finding catalog type by name: %s", err)
	}

	for _, v := range catalogTypes {
		if v.Name == targetName {
			return &v, nil
		}
	}

	return nil, errors.Errorf("catalog type %q not found", targetName)
}

func FindCatalogEntryByNameWithTypeName(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, targetName string, typeName string) (*client.CatalogEntryV2, error) {
	catalogType, err := FindCatalogTypeByName(ctx, logger, cl, typeName)
	if err != nil {
		return nil, fmt.Errorf("finding catalog type by name for entry lookup: %s", err)
	}

	catalogEntry, err := FindCatalogEntryByNameWithTypeID(ctx, logger, cl, targetName, catalogType.Id)
	if err != nil {
		return nil, fmt.Errorf("finding catalog entry by name with type id: %s", err)
	}

	return catalogEntry, nil
}

func FindCatalogEntryByNameWithTypeID(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, targetName string, typeID string) (*client.CatalogEntryV2, error) {
	var (
		after    *string
		pageSize = 250
	)

	for {
		page, err := cl.CatalogV2ListEntriesWithResponse(ctx, &client.CatalogV2ListEntriesParams{
			CatalogTypeId: typeID,
			PageSize:      &pageSize,
			After:         after,
		})
		if err != nil {
			return nil, fmt.Errorf("listing catalog entries: %s", err)
		}

		for _, candidate := range page.JSON200.CatalogEntries {
			if candidate.Name == targetName {
				return &candidate, nil
			}
			for _, alias := range candidate.Aliases {
				if alias == targetName {
					return &candidate, nil
				}
			}
		}

		if count := len(page.JSON200.CatalogEntries); count == 0 {
			return nil, fmt.Errorf("failed to find catalog entry %q", targetName) // end pagination
		} else {
			after = lo.ToPtr(page.JSON200.CatalogEntries[count-1].Id)
		}
	}
}

func FindCatalogEntryByID(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, targetID string) (*client.CatalogEntryV2, error) {
	res, err := cl.CatalogV2ShowEntryWithResponse(ctx, targetID)
	if err != nil {
		return nil, errors.Wrap(err, "listing catalog")
	}
	return &res.JSON200.CatalogEntry, nil
}

func ListAllCustomFields(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses) ([]client.CustomFieldV2, error) {
	res, err := cl.CustomFieldsV2ListWithResponse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "listing catalog")
	}
	return res.JSON200.CustomFields, nil
}

func ListAllCatalogTypes(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses) ([]client.CatalogTypeV2, error) {
	res, err := cl.CatalogV2ListTypesWithResponse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "listing catalog types")
	}
	return res.JSON200.CatalogTypes, nil
}

func ListAllCatalogEntries(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses) ([]client.CatalogEntryV2, error) {
	catalogTypes, err := ListAllCatalogTypes(ctx, logger, cl)
	if err != nil {
		return nil, fmt.Errorf("failed enumerating catalog types: %s", err)
	}

	var results []client.CatalogEntryV2

	for _, catalogType := range catalogTypes {
		var (
			after    *string
			pageSize = 250
		)

		for {
			page, err := cl.CatalogV2ListEntriesWithResponse(ctx, &client.CatalogV2ListEntriesParams{
				CatalogTypeId: catalogType.Id,
				PageSize:      &pageSize,
				After:         after,
			})
			if err != nil {
				return nil, fmt.Errorf("listing catalog entries: %q", err)
			}

			results = append(results, page.JSON200.CatalogEntries...)

			if count := len(page.JSON200.CatalogEntries); count == 0 {
				break // end pagination
			} else {
				after = lo.ToPtr(page.JSON200.CatalogEntries[count-1].Id)
			}
		}
	}

	return results, nil
}

func ListAllCatalogEntriesByTypeName(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, typeName string) ([]client.CatalogEntryV2, error) {
	catalogType, err := FindCatalogTypeByName(ctx, logger, cl, typeName)
	if err != nil {
		return nil, fmt.Errorf("finding catalog type by name for entry lookup: %s", err)
	}

	return ListAllCatalogEntriesByTypeID(ctx, logger, cl, catalogType.Id)
}

func ListAllCatalogEntriesByTypeID(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, typeID string) ([]client.CatalogEntryV2, error) {
	var (
		after    *string
		pageSize = 250
		results  = []client.CatalogEntryV2{}
	)

	for {
		page, err := cl.CatalogV2ListEntriesWithResponse(ctx, &client.CatalogV2ListEntriesParams{
			CatalogTypeId: typeID,
			PageSize:      &pageSize,
			After:         after,
		})
		if err != nil {
			return nil, fmt.Errorf("listing catalog entries: %q", err)
		}

		results = append(results, page.JSON200.CatalogEntries...)

		if count := len(page.JSON200.CatalogEntries); count == 0 {
			return results, nil // end pagination
		} else {
			after = lo.ToPtr(page.JSON200.CatalogEntries[count-1].Id)
		}
	}
}

func EditIncidentByReferenceNumber(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, reference int, newCustomFields map[string]string) (*client.IncidentV2, error) {
	incident, err := FindIncidentByReferenceNumber(ctx, logger, cl, reference)
	if err != nil {
		return nil, fmt.Errorf("finding incident by id to show: %q", err)
	}

	return EditIncident(ctx, logger, cl, incident.Id, newCustomFields)
}

func ShowIncidentByID(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, id string) (*client.IncidentV2, error) {
	res, err := cl.IncidentsV2ShowWithResponse(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "listing catalog types")
	}
	return &res.JSON200.Incident, nil
}

func ShowIncidentByReference(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, reference int) (*client.IncidentV2, error) {
	incident, err := FindIncidentByReferenceNumber(ctx, logger, cl, reference)
	if err != nil {
		return nil, errors.Wrap(err, "finding incident by id to show")
	}

	return ShowIncidentByID(ctx, logger, cl, incident.Id)
}

func ListAllIncidents(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses) ([]client.IncidentV2, error) {
	var (
		after    *string
		pageSize = int64(250)
		results  = []client.IncidentV2{}
	)

	for {
		page, err := cl.IncidentsV2ListWithResponse(ctx, &client.IncidentsV2ListParams{
			PageSize: &pageSize,
			After:    after,
		})
		if err != nil {
			return nil, errors.Wrap(err, "listing incidents")
		}

		results = append(results, page.JSON200.Incidents...)

		if count := len(page.JSON200.Incidents); count == 0 {
			return results, nil // end pagination
		} else {
			after = lo.ToPtr(page.JSON200.Incidents[count-1].Id)
		}
	}
}

func FindIncidentByReferenceNumber(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, reference int) (*client.IncidentV2, error) {
	incidents, err := ListAllIncidents(ctx, logger, cl)
	if err != nil {
		return nil, errors.Wrap(err, "listing incidents to find by reference")
	}

	for _, candidate := range incidents {
		if candidate.Reference == fmt.Sprintf("INC-%d", reference) {
			return &candidate, nil
		}
	}

	return nil, errors.New("incident not found")
}

func EditIncident(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses, id string, newCustomFields map[string]string) (*client.IncidentV2, error) {
	customFields, err := ListAllCustomFields(ctx, logger, cl)
	if err != nil {
		return nil, fmt.Errorf("failed to find custom field types: %q", err)
	}

	customFieldIDToType := map[string]client.CustomFieldV2FieldType{}
	customFieldMap := map[string]string{}
	for _, candidate := range customFields {
		customFieldMap[candidate.Name] = candidate.Id
		customFieldIDToType[candidate.Id] = candidate.FieldType
	}

	body := client.IncidentsV2EditJSONRequestBody{
		Incident: client.IncidentEditPayloadV2{
			CustomFieldEntries: &[]client.CustomFieldEntryPayloadV1{},
		},
		NotifyIncidentChannel: false,
	}

	for k, v := range newCustomFields {
		id, ok := customFieldMap[k]
		if !ok {
			return nil, errors.Errorf("custom field ID for %q not found", k)
		}

		// empty array resets previous set value
		var values = []client.CustomFieldValuePayloadV1{}

		if v != "" {
			var customVal client.CustomFieldValuePayloadV1
			switch customFieldIDToType[id] {
			case client.SingleSelect:
				customVal = client.CustomFieldValuePayloadV1{
					ValueCatalogEntryId: &v,
				}

			case client.Text:
				customVal = client.CustomFieldValuePayloadV1{
					ValueText: &v,
				}
			default:
				return nil, errors.Errorf("unsupported custom field type %q", customFieldIDToType[id])
			}

			values = append(values, customVal)
		}

		*body.Incident.CustomFieldEntries = append(*body.Incident.CustomFieldEntries, client.CustomFieldEntryPayloadV1{
			CustomFieldId: id,
			Values:        values,
		})
	}

	litter.Dump(id)
	litter.Dump("===")
	litter.Dump(body)

	res, err := cl.IncidentsV2EditWithResponse(ctx, id, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to edit incident")
	}
	return &res.JSON200.Incident, nil
}
