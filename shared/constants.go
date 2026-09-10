package utils

const (
	DEVICES_FINDONE      = "devices.findone"
	DEVICES_FINDALL      = "devices.findall"
	DEVICES_ADD          = "devices.add"
	DEVICES_UPDATE       = "devices.update"
	DEVICES_UPDATE_ONE   = "devices.updateone"
	DEVICES_UPDATE_OWNER = "devices.updateowner"
	DEVICES_PANIC        = "devices.panic"
	DEVICES_IMPORT       = "devices.import"
	DEVICES_GEOCODE      = "devices.geocode"
	DEVICES_PROVISION    = "devices.provision"
	DEVICES_SET_ODO      = "devices.setodo"
	DEVICES_IMMOBILIZE   = "devices.immobilize"
	DEVICES_STATUS       = "devices.status"
	DEVICES_TRANSFER     = "devices.transfer"
)

const (
	DEVICE_ACTIVE    = "ACTIVE"
	DEVICE_SUSPENDED = "SUSPENDED"
	DEVICE_DELETED   = "DELETED"
)

const (
	EMAIL_SEND = "email.send"
)

const (
	ALERT_FIND      = "alert.find"
	ALERT_FIND_ALL  = "alert.findall"
	ALERT_ADD_EVENT = "alert.addevent"
	ALERT_ADD       = "alert.add"
)

const (
	MOVING        = "MOVING"
	STOPPED       = "STOPPED"
	PARKED        = "PARKED"
	IDLE          = "IDLE"
	NOT_STRACKING = "NOT_TRACKING"
)

const (
	SMS_SEND      = "sms.send"
	WHATSAPP_SEND = "whatsapp.send"
)

const (
	REPORTS_FIND_ALL        = "blueasset.reports.findall"
	REPORTS_SCHEDULE_ADD    = "blueasset.reports.schedule.add"
	REPORTS_SCHEDULE_UPDATE = "blueasset.reports.schedule.update"
	REPORTS_SCHEDULE_DELETE = "blueasset.reports.schedule.delete"
	REPORTS_SCHEDULE_FIND   = "blueasset.reports.schedule.find"
	REPORTS_SCHEDULE_RUN    = "blueasset.reports.schedule.run"
	REPORTS_PARAM_OPTIONS   = "blueasset.reports.param.options"
)

const (
	VAS_ADD                        = "vas.add"
	VAS_UPDATE                     = "vas.update"
	VAS_FINDALL                    = "vas.findall"
	VAS_FIND                       = "vas.find"
	VAS_SUBSCRIBE                  = "vas.subscriptions.add"
	VAS_FIND_SUBSCRIPTIONS         = "vas.subscriptions.findall"
	VAS_FIND_SUBSCRIPTION          = "vas.subscriptions.find"
	VAS_UPDATE_SUBSCRIPTION        = "vas.subscriptions.update"
	VAS_UPDATE_SUBSCRIPTIONS       = "vas.subscriptions.updatebulk"
	VAS_CLONE_SUBSCRIPTIONS        = "vas.subscriptions.clone"
	VAS_UPDATE_SUBSCRIPTION_STATUS = "vas.subscriptions.update.status"
)

const (
	DEVICES_TYPE_FINDALL           = "devices.types.findall"
	DEVICES_TYPE_ADD               = "devices.types.add"
	DEVICES_CONFIG_FINDALL         = "devices.configs.findall"
	DEVICES_CONFIG_VERSION_FINDALL = "devices.versions.findall"
)

const (
	INSTALL_ADD    = "install.add"
	INSTALL_UPLOAD = "install.upload"
)

const (
	PROFILES_LOGIN    = "profiles.login"
	PROFILES_LOGIN_AS = "profiles.loginas"
	PROFILES_ADD      = "profiles.add"
	PROFILES_FINDALL  = "profiles.findall"
	PROFILES_FIND     = "profiles.find"
	PROFILES_UPDATE   = "profiles.update"
	//PROFILE_UPLOADS_GET     = "profiles.uploads.get"
	//PROFILE_UPLOADS_DELETE  = "profiles.uploads.delete"
	//PROFILE_AVATAR_ADD      = "profiles.avatar.add"
	PROFILE_CHANGE_PASSWORD  = "profiles.changepassword"
	PROFILE_FORGET_PASSWORD  = "profiles.forgetpassword"
	PROFILE_CHANGE_PASSWORD2 = "profiles.changepassword2"

	//PROFILE_AVATAR_GET      = "profiles.avatar.get"

	PROFILE_SUB_PROFILE_ADD     = "profiles.subprofiles.add"
	PROFILE_SUB_PROFILE_FINDALL = "profiles.subprofiles.findall"
	PROFILE_SUB_PROFILE_REMOVE  = "profiles.subprofiles.remove"

	PROFILE_ROLES_FIND_ALL   = "profiles.roles.findall"
	PROFILE_DEVICES_FIND_ALL = "profiles.devices.findall"
	PROFILE_DEVICES_ADD      = "profiles.devices.add"
	PROFILE_DEVICES_REMOVE   = "profiles.devices.remove"

	PROFILES_DEVICES_GROUPS_ADD    = "profiles.devices.groups.add"
	PROFILES_DEVICES_GROUPS_FIND   = "profiles.devices.groups.find"
	PROFILES_DEVICES_GROUPS_UPDATE = "profiles.devices.groups.update"
	PROFILES_DEVICES_GROUPS_DELETE = "profiles.devices.groups.remove"
)

const (
	PROFILES_CONTACTS_GROUP_FINDALL           = "profiles.contacts.groups.findall"
	PROFILES_CONTACTS_GROUP_ADD               = "profiles.contacts.groups.add"
	PROFILES_CONTACTS_GROUP_UPDATE            = "profiles.contacts.groups.update"
	PROFILES_CONTACTS_GROUP_DELETE            = "profiles.contacts.groups.delete"
	PROFILES_CONTACTS_GROUP_GET               = "profiles.contacts.groups.find"
	PROFILES_CONTACTS_GROUP_ADD_TO_GROUP      = "profiles.contacts.groups.addcontacts"
	PROFILES_CONTACTS_GROUP_REMOVE_FROM_GROUP = "profiles.contacts.groups.removecontacts"
)

const (
	PROFILE_ACTIVE    = "active"
	PROFILE_SUSPENDED = "suspended"
	PROFILE_DELETED   = "deleted"
	PROFILE_INACTIVE  = "inactive"
)
const (
	PROFILES_CONTACTS_FINDALL = "profiles.contacts.findall"
	PROFILES_CONTACTS_FIND    = "profiles.contacts.find"
	PROFILES_CONTACTS_ADD     = "profiles.contacts.add"
	PROFILES_CONTACTS_DELETE  = "profiles.contacts.remove"
	PROFILES_CONTACTS_UPDATE  = "profiles.contacts.update"
)

const (
	HISTORY_RAW       = "history.raw"
	HISTORY_LAST      = "history.last"
	HISTORY_TRIPS     = "history.trips"
	HISTORY_TRIP_SAVE = "history.trips.save"
	HISTORY_POINT     = "history.point"
)

const (
	GEOFENCE_ADD        = "geofence.add"
	GEOFENCE_UPDATE     = "geofence.update"
	GEOFENCE_ADD_DEVICE = "geofence.add.device"
	GEOFENCE_FINDALL    = "geofence.findall"
	GEOFENCE_FIND       = "geofence.find"
	GEOFENCE_DELETE     = "geofence.remove"
	GEOFENCE_INSIDE     = "geofence.inside"
)

const (
	TASKS_ADD = "tasks.add"
)
