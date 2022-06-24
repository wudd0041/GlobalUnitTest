package constraint

var keys_default_value map[string]interface{}

func init() {
	keys_default_value = map[string]interface{}{
		"ones:app:project:enable":     false,
		"ones:app:wiki:enable":        false,
		"ones:app:testcase:enable":    false,
		"ones:app:pipeline:enable":    false,
		"ones:app:performance:enable": false,
		"ones:app:plan:enable":        false,
		"ones:app:account:enable":     false,
		"ones:app:desk:enable":        false,
		"ones:app:automation:enable":  false,
	}

	// for mock
	keys_mock_value := map[string]interface{}{
		"ones:app:project:ppm":                     false,
		"ones:app:project:ppm:plan":                false,
		"ones:app:project:ppm:deliverable":         false,
		"ones:app:project:ppm:milestone":           false,
		"ones:app:project:custom_object_link_type": false,
		"ones:app:project:workflow:field":          false,
		"ones:app:project:workflow:post_function":  false,
		"ones:app:testcase:custom_report_template": false,
		"ones:file_capacity":                       1024 * 1024 * 1024 * 5,
	}
	for k, v := range keys_mock_value {
		keys_default_value[k] = v
	}
}
