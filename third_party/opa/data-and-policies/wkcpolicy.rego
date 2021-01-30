package wkcpolicy
import data.data_policies as dp

transform[action] {
	description = "Columns with Confidential tag to be redacted before read"
	dp.correct_input
    #user context and access type check
    #dp.check_access_type([dp.AccessTypes.READ])
	dp.check_access_type(["READ"])
	dp.check_purpose("Fraud Detection")
	dp.check_role("Data Scientist")
	dp.dataset_has_tag("residency = Turkey")
    column_names := dp.column_with_tag("Confidential")
    action = dp.build_redact_column_action(column_names[_], dp.build_policy_from_description(description))
}

deny[action] {
	description = "deny if role is not Data Scientist when purpose is Fraud Detection"
	#dp.correct_input
    #user context and access type check
    #dp.check_access_type([dp.AccessTypes.READ])
	dp.check_access_type(["READ"])
	dp.check_purpose("Fraud Detection")
	dp.check_role_not("Data Scientist")
	dp.dataset_has_tag("residency = Turkey")
    action = dp.build_deny_access_action(dp.build_policy_from_description(description))
}

deny[action] {
	description = "If Columns with Confidential tag deny access"
	#dp.correct_input
    #user context and access type check
    #dp.check_access_type([dp.AccessTypes.READ])
	dp.check_access_type(["READ"])
	dp.check_purpose("Customer Behaviour Analysis")
	dp.check_role("Business Analyst")
	dp.dataset_has_tag("residency = Turkey")
    dp.column_has_tag("Confidential")
    action = dp.build_deny_access_action(dp.build_policy_from_description(description))
}

deny[action] {
	description = "deny if role is not Business Analyst when purpose is Customer Behaviour Analysis"
	#dp.correct_input
    #user context and access type check
    #dp.check_access_type([dp.AccessTypes.READ])
	dp.check_access_type(["READ"])
	dp.check_purpose("Customer Behaviour Analysis")
	dp.check_role_not("Business Analyst")
	dp.dataset_has_tag("residency = Turkey")
    action = dp.build_deny_access_action(dp.build_policy_from_description(description))
}

deny[action] {
	description = "If data residency is Turkey but processing geography is not Turkey then deny writing"
	#dp.correct_input
    #user context and access type check
    #dp.check_access_type([dp.AccessTypes.WRITE])
	dp.check_access_type(["WRITE"])
	dp.dataset_has_tag("residency = Turkey")
	dp.check_processingGeo_not("Turkey")
    action = dp.build_deny_write_action(dp.build_policy_from_description(description))
}

deny[action] {
	description = "If data residency is not Turkey but processing geography is not Turkey or EEA then deny writing"
	#dp.correct_input
    #user context and access type check
    #dp.check_access_type([dp.AccessTypes.WRITE])
	dp.check_access_type(["WRITE"])
	dp.dataset_has_tag_not("residency = Turkey")
	dp.check_processingGeo_not("Turkey")
	dp.check_processingGeo_not("EEA")
    action = dp.build_deny_write_action(dp.build_policy_from_description(description))
}
