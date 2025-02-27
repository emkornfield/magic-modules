package google

import (
	"fmt"
	"strings"

	transport_tpg "github.com/GoogleCloudPlatform/terraform-google-conversion/v2/tfplan2cai/converters/google/resources/transport"
)

func resourceConverterOrgPolicyPolicy() ResourceConverter {
	return ResourceConverter{
		Convert:           GetV2OrgPoliciesCaiObject,
		MergeCreateUpdate: MergeV2OrgPolicies,
	}
}

func GetV2OrgPoliciesCaiObject(d TerraformResourceData, config *transport_tpg.Config) ([]Asset, error) {
	assetNamePattern, assetType, err := getAssetNameAndTypeFromParent(d.Get("parent").(string))
	if err != nil {
		return []Asset{}, err
	}

	name, err := assetName(d, config, assetNamePattern)
	if err != nil {
		return []Asset{}, err
	}

	if obj, err := GetV2OrgPoliciesApiObject(d, config); err == nil {
		return []Asset{{
			Name:          name,
			Type:          assetType,
			V2OrgPolicies: []*V2OrgPolicies{&obj},
		}}, nil
	} else {
		return []Asset{}, err
	}

}

func GetV2OrgPoliciesApiObject(d TerraformResourceData, config *transport_tpg.Config) (V2OrgPolicies, error) {
	spec, err := expandSpecV2OrgPolicies(d.Get("spec").([]interface{}))
	if err != nil {
		return V2OrgPolicies{}, err
	}

	return V2OrgPolicies{
		Name:       d.Get("name").(string),
		PolicySpec: spec,
	}, nil
}

func MergeV2OrgPolicies(existing, incoming Asset) Asset {
	existing.Resource = incoming.Resource
	return existing
}

func getAssetNameAndTypeFromParent(parent string) (assetName string, assetType string, err error) {
	const prefix = "cloudresourcemanager.googleapis.com/"
	if strings.Contains(parent, "projects") {
		return "//" + prefix + parent, prefix + "Project", nil
	} else if strings.Contains(parent, "folders") {
		return "//" + prefix + parent, prefix + "Folder", nil
	} else if strings.Contains(parent, "organizations") {
		return "//" + prefix + parent, prefix + "Organization", nil
	} else {
		return "", "", fmt.Errorf("Invalid parent address(%s) for an asset", parent)
	}
}

func expandSpecV2OrgPolicies(configured []interface{}) (*PolicySpec, error) {
	if len(configured) == 0 || configured[0] == nil {
		return nil, nil
	}

	specMap := configured[0].(map[string]interface{})

	policyRules, err := expandPolicyRulesSpec(specMap["rules"].([]interface{}))
	if err != nil {
		return &PolicySpec{}, err
	}

	return &PolicySpec{
		Etag:              specMap["etag"].(string),
		PolicyRules:       policyRules,
		InheritFromParent: specMap["inherit_from_parent"].(bool),
		Reset:             specMap["reset"].(bool),
	}, nil

}

func expandPolicyRulesSpec(configured []interface{}) ([]*PolicyRule, error) {
	if len(configured) == 0 || configured[0] == nil {
		return nil, nil
	}

	var policyRules []*PolicyRule
	for i := 0; i < len(configured); i++ {
		policyRule, err := expandPolicyRulePolicyRules(configured[i])
		if err != nil {
			return nil, err
		}
		policyRules = append(policyRules, policyRule)
	}

	return policyRules, nil

}

func expandPolicyRulePolicyRules(configured interface{}) (*PolicyRule, error) {
	policyRuleMap := configured.(map[string]interface{})

	values, err := expandValuesPolicyRule(policyRuleMap["values"].([]interface{}))
	if err != nil {
		return &PolicyRule{}, err
	}

	allowAll, err := convertStringToBool(policyRuleMap["allow_all"].(string))
	if err != nil {
		return &PolicyRule{}, err
	}

	denyAll, err := convertStringToBool(policyRuleMap["deny_all"].(string))
	if err != nil {
		return &PolicyRule{}, err
	}

	enforce, err := convertStringToBool(policyRuleMap["enforce"].(string))
	if err != nil {
		return &PolicyRule{}, err
	}

	condition, err := expandConditionPolicyRule(policyRuleMap["condition"].([]interface{}))
	if err != nil {
		return &PolicyRule{}, err
	}
	return &PolicyRule{
		Values:    values,
		AllowAll:  allowAll,
		DenyAll:   denyAll,
		Enforce:   enforce,
		Condition: condition,
	}, nil
}

func expandValuesPolicyRule(configured []interface{}) (*StringValues, error) {
	if len(configured) == 0 || configured[0] == nil {
		return nil, nil
	}
	valuesMap := configured[0].(map[string]interface{})
	return &StringValues{
		AllowedValues: convertInterfaceToStringArray(valuesMap["allowed_values"].([]interface{})),
		DeniedValues:  convertInterfaceToStringArray(valuesMap["denied_values"].([]interface{})),
	}, nil
}

func expandConditionPolicyRule(configured []interface{}) (*Expr, error) {
	if len(configured) == 0 || configured[0] == nil {
		return nil, nil
	}
	conditionMap := configured[0].(map[string]interface{})
	return &Expr{
		Expression:  conditionMap["expression"].(string),
		Title:       conditionMap["title"].(string),
		Description: conditionMap["description"].(string),
		Location:    conditionMap["location"].(string),
	}, nil
}

func convertStringToBool(val string) (bool, error) {
	if (val == "false") || (val == "FALSE") || (val == "") {
		return false, nil
	} else if (val == "true") || (val == "TRUE") {
		return true, nil
	}

	return false, fmt.Errorf("Invalid value for a boolean field: %s", val)
}

func convertInterfaceToStringArray(values []interface{}) []string {
	stringArray := make([]string, len(values))
	for i, v := range values {
		stringArray[i] = v.(string)
	}
	return stringArray
}
