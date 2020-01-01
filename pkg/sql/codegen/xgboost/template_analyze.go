// Copyright 2019 The SQLFlow Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xgboost

import (
	"text/template"
)

type analyzeFiller struct {
	DataSource           string
	DatasetSQL           string
	ShapSummaryParams    string
	FeatureFieldMetaJSON string
	LabelName            string
}

const analyzeTemplateText = `
import json
from sqlflow_submitter.xgboost.explain import explain

feature_field_meta = json.loads('''{{.FeatureFieldMetaJSON}}''')
label_name = '''{{.LabelName}}'''
summary_params = json.loads('''{{.ShapSummaryParams}}''')

explain(
    datasource='''{{.DataSource}}''',
    select='''{{.DatasetSQL}}''',
    feature_field_meta=feature_field_meta,
    label_name=label_name,
    summary_params=summary_params)
`

var analyzeTemplate = template.Must(template.New("analyze").Parse(analyzeTemplateText))
