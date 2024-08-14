package main

import (
	"bytes"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

// http://swagger.io/specification/#infoObject
type swaggerInfoObject struct {
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	TermsOfService string `json:"termsOfService,omitempty"`
	Version        string `json:"version"`

	Contact *swaggerContactObject `json:"contact,omitempty"`
	License *swaggerLicenseObject `json:"license,omitempty"`
}

// http://swagger.io/specification/#contactObject
type swaggerContactObject struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// http://swagger.io/specification/#licenseObject
type swaggerLicenseObject struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// http://swagger.io/specification/#externalDocumentationObject
type swaggerExternalDocumentationObject struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

// http://swagger.io/specification/#swaggerObject
type swaggerObject struct {
	Swagger             string                              `json:"swagger"`
	Info                swaggerInfoObject                   `json:"info"`
	Host                string                              `json:"host,omitempty"`
	BasePath            string                              `json:"basePath,omitempty"`
	Schemes             []string                            `json:"schemes"`
	Consumes            []string                            `json:"consumes"`
	Produces            []string                            `json:"produces"`
	Paths               swaggerPathsObject                  `json:"paths"`
	Definitions         swaggerDefinitionsObject            `json:"definitions"`
	StreamDefinitions   swaggerDefinitionsObject            `json:"x-stream-definitions,omitempty"`
	SecurityDefinitions swaggerSecurityDefinitionsObject    `json:"securityDefinitions,omitempty"`
	Security            []swaggerSecurityRequirementObject  `json:"security,omitempty"`
	ExternalDocs        *swaggerExternalDocumentationObject `json:"externalDocs,omitempty"`
}

// http://swagger.io/specification/#securityDefinitionsObject
type swaggerSecurityDefinitionsObject map[string]swaggerSecuritySchemeObject

// http://swagger.io/specification/#securitySchemeObject
type swaggerSecuritySchemeObject struct {
	Type             string              `json:"type"`
	Description      string              `json:"description,omitempty"`
	Name             string              `json:"name,omitempty"`
	In               string              `json:"in,omitempty"`
	Flow             string              `json:"flow,omitempty"`
	AuthorizationURL string              `json:"authorizationUrl,omitempty"`
	TokenURL         string              `json:"tokenUrl,omitempty"`
	Scopes           swaggerScopesObject `json:"scopes,omitempty"`
}

// http://swagger.io/specification/#scopesObject
type swaggerScopesObject map[string]string

// http://swagger.io/specification/#securityRequirementObject
type swaggerSecurityRequirementObject map[string][]string

// http://swagger.io/specification/#pathsObject
type swaggerPathsObject map[string]swaggerPathItemObject

type swaggerPathSummary struct {
	Path    string
	Summary string
	Tags    string
}

func (po swaggerPathsObject) MarshalJSON() ([]byte, error) {
	psv := []swaggerPathSummary{}

	for k, v := range po {
		if len(k) == 0 {
			continue
		}
		summary := k
		tags := k
		if v.Get != nil {
			if len(v.Get.Summary) > 0 {
				summary = v.Get.Summary
			}

			if len(v.Get.Tags) > 0 && len(v.Get.Tags[0]) > 0 {
				tags = v.Get.Tags[0]
			}

		} else if v.Post != nil {
			if len(v.Post.Summary) > 0 {
				summary = v.Post.Summary
			}

			if len(v.Post.Tags) > 0 && len(v.Post.Tags[0]) > 0 {
				tags = v.Post.Tags[0]
			}
		} else if v.Delete != nil {
			if len(v.Delete.Summary) > 0 {
				summary = v.Delete.Summary
			}

			if len(v.Delete.Tags) > 0 && len(v.Delete.Tags[0]) > 0 {
				tags = v.Delete.Tags[0]
			}

		} else if v.Put != nil {
			if len(v.Put.Summary) > 0 {
				summary = v.Put.Summary
			}

			if len(v.Put.Tags) > 0 && len(v.Put.Tags[0]) > 0 {
				tags = v.Put.Tags[0]
			}
		} else if v.Patch != nil {
			if len(v.Patch.Summary) > 0 {
				summary = v.Patch.Summary
			}

			if len(v.Patch.Tags) > 0 && len(v.Patch.Tags[0]) > 0 {
				tags = v.Patch.Tags[0]
			}
		} else {
			summary = k
			tags = k
		}

		psv = append(psv, swaggerPathSummary{
			Path:    k,
			Summary: summary,
			Tags:    tags,
		})
	}

	sort.Slice(psv, func(i, j int) bool {
		rsi := strings.FieldsFunc(psv[i].Summary, func(r rune) bool {
			return strings.ContainsRune(".,。:", r)
		})

		if len(rsi) == 0 {
			return false
		}

		rsj := strings.FieldsFunc(psv[j].Summary, func(r rune) bool {
			return strings.ContainsRune(".,。:", r)
		})

		if len(rsj) == 0 {
			return true
		}

		rti := strings.FieldsFunc(psv[i].Tags, func(r rune) bool {
			return strings.ContainsRune(".,。:", r)
		})

		if len(rti) == 0 {
			return false
		}

		rtj := strings.FieldsFunc(psv[j].Tags, func(r rune) bool {
			return strings.ContainsRune(".,。:", r)
		})

		if len(rtj) == 0 {
			return true
		}

		indexsi, errsi := strconv.Atoi(strings.TrimSpace(rsi[0]))
		indexsj, errsj := strconv.Atoi(strings.TrimSpace(rsj[0]))

		indexti, errti := strconv.Atoi(strings.TrimSpace(rti[0]))
		indextj, errtj := strconv.Atoi(strings.TrimSpace(rtj[0]))

		if errti == nil && errtj == nil {
			if indexti == indextj {
				if errsi == nil && errsj == nil {
					return indexsi < indexsj
				} else if errsi != nil {
					return false
				} else if errsj != nil {
					return true
				} else {
					return psv[i].Summary < psv[j].Summary
				}
			} else {
				return indexti < indextj
			}

		} else if errti != nil {
			return false
		} else if errtj != nil {
			return true
		} else {
			return psv[i].Tags < psv[j].Tags
		}
	})

	var buf bytes.Buffer
	buf.WriteString("{")
	for i, ps := range psv {
		if len(ps.Path) == 0 {
			continue
		}

		if i != 0 {
			buf.WriteString(",")
		}

		path, err := json.Marshal(ps.Path)
		if err != nil {
			return nil, err
		}

		buf.Write(path)
		buf.WriteString(":")
		po, _ := po[ps.Path]
		val, err := json.Marshal(po)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}
	buf.WriteString("}")

	return buf.Bytes(), nil
}

// http://swagger.io/specification/#pathItemObject
type swaggerPathItemObject struct {
	Get    *swaggerOperationObject `json:"get,omitempty"`
	Delete *swaggerOperationObject `json:"delete,omitempty"`
	Post   *swaggerOperationObject `json:"post,omitempty"`
	Put    *swaggerOperationObject `json:"put,omitempty"`
	Patch  *swaggerOperationObject `json:"patch,omitempty"`
}

// http://swagger.io/specification/#operationObject
type swaggerOperationObject struct {
	Summary     string                  `json:"summary,omitempty"`
	Description string                  `json:"description,omitempty"`
	OperationID string                  `json:"operationId,omitempty"`
	Responses   swaggerResponsesObject  `json:"responses"`
	Parameters  swaggerParametersObject `json:"parameters,omitempty"`
	Tags        []string                `json:"tags,omitempty"`
	Deprecated  bool                    `json:"deprecated,omitempty"`

	Security     *[]swaggerSecurityRequirementObject `json:"security,omitempty"`
	ExternalDocs *swaggerExternalDocumentationObject `json:"externalDocs,omitempty"`
}

type swaggerParametersObject []swaggerParameterObject

// http://swagger.io/specification/#parameterObject
type swaggerParameterObject struct {
	Name             string              `json:"name"`
	Description      string              `json:"description,omitempty"`
	In               string              `json:"in,omitempty"`
	Required         bool                `json:"required"`
	Type             string              `json:"type,omitempty"`
	Format           string              `json:"format,omitempty"`
	Items            *swaggerItemsObject `json:"items,omitempty"`
	Enum             []string            `json:"enum,omitempty"`
	CollectionFormat string              `json:"collectionFormat,omitempty"`
	Default          string              `json:"default,omitempty"`
	MinItems         *int                `json:"minItems,omitempty"`

	// Or you can explicitly refer to another type. If this is defined all
	// other fields should be empty
	Schema *swaggerSchemaObject `json:"schema,omitempty"`
}

// core part of schema, which is common to itemsObject and schemaObject.
// http://swagger.io/specification/#itemsObject
type schemaCore struct {
	Type    string          `json:"type,omitempty"`
	Format  string          `json:"format,omitempty"`
	Ref     string          `json:"$ref,omitempty"`
	Example json.RawMessage `json:"example,omitempty"`

	Items *swaggerItemsObject `json:"items,omitempty"`

	// If the item is an enumeration include a list of all the *NAMES* of the
	// enum values.  I'm not sure how well this will work but assuming all enums
	// start from 0 index it will be great. I don't think that is a good assumption.
	Enum    []string `json:"enum,omitempty"`
	Default string   `json:"default,omitempty"`
}

type swaggerItemsObject schemaCore

func (o *swaggerItemsObject) getType() string {
	if o == nil {
		return ""
	}
	return o.Type
}

// http://swagger.io/specification/#responsesObject
type swaggerResponsesObject map[string]swaggerResponseObject

// http://swagger.io/specification/#responseObject
type swaggerResponseObject struct {
	Description string              `json:"description"`
	Schema      swaggerSchemaObject `json:"schema"`
}

type keyVal struct {
	Key   string
	Value interface{}
}

type swaggerSchemaObjectProperties []keyVal

func (op swaggerSchemaObjectProperties) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, kv := range op {
		if i != 0 {
			buf.WriteString(",")
		}
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		val, err := json.Marshal(kv.Value)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

// http://swagger.io/specification/#schemaObject
type swaggerSchemaObject struct {
	schemaCore
	// Properties can be recursively defined
	Properties           *swaggerSchemaObjectProperties `json:"properties,omitempty"`
	AdditionalProperties *swaggerSchemaObject           `json:"additionalProperties,omitempty"`

	Description string `json:"description,omitempty"`
	Title       string `json:"title,omitempty"`

	ExternalDocs *swaggerExternalDocumentationObject `json:"externalDocs,omitempty"`

	MultipleOf       float64  `json:"multipleOf,omitempty"`
	Maximum          float64  `json:"maximum,omitempty"`
	ExclusiveMaximum bool     `json:"exclusiveMaximum,omitempty"`
	Minimum          float64  `json:"minimum,omitempty"`
	ExclusiveMinimum bool     `json:"exclusiveMinimum,omitempty"`
	MaxLength        uint64   `json:"maxLength,omitempty"`
	MinLength        uint64   `json:"minLength,omitempty"`
	Pattern          string   `json:"pattern,omitempty"`
	MaxItems         uint64   `json:"maxItems,omitempty"`
	MinItems         uint64   `json:"minItems,omitempty"`
	UniqueItems      bool     `json:"uniqueItems,omitempty"`
	MaxProperties    uint64   `json:"maxProperties,omitempty"`
	MinProperties    uint64   `json:"minProperties,omitempty"`
	Required         []string `json:"required,omitempty"`
}

// http://swagger.io/specification/#definitionsObject
type swaggerDefinitionsObject map[string]swaggerSchemaObject
