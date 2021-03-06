package common

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
)

var ApiTlsMinVersion uint16 = tls.VersionTLS12
var ApiTlsCipherSuites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
}

type apiBadHttpBody struct {
	body   []byte
	reason string
}

func (self *apiBadHttpBody) Error() string {
	return fmt.Sprintf("bad HTTP body %#v: %s", string(self.body), self.reason)
}

var pkgMgrActions2api = map[PkgMgrAction]string{
	PkgMgrInstall:   "install",
	PkgMgrUpdate:    "update",
	PkgMgrConfigure: "configure",
	PkgMgrRemove:    "remove",
	PkgMgrPurge:     "purge",
}

var api2pkgMgrActions = map[string]PkgMgrAction{
	"install":   PkgMgrInstall,
	"update":    PkgMgrUpdate,
	"configure": PkgMgrConfigure,
	"remove":    PkgMgrRemove,
	"purge":     PkgMgrPurge,
}

func PkgMgrTasks2Api(tasks map[PkgMgrTask]struct{}) (jsn []byte, err error) {
	apiTasks := make([]interface{}, len(tasks))
	apiTaskIdx := 0

	for task := range tasks {
		record := map[string]interface{}{
			"package": task.PackageName,
			"action":  pkgMgrActions2api[task.Action],
		}

		if task.FromVersion != "" {
			record["from_version"] = task.FromVersion
		}

		if task.ToVersion != "" {
			record["to_version"] = task.ToVersion
		}

		apiTasks[apiTaskIdx] = record
		apiTaskIdx++
	}

	return json.Marshal(apiTasks)
}

func Api2PkgMgrTasks(body []byte) (tasks map[PkgMgrTask]struct{}, err error) {
	var unJson interface{}
	if errJU := json.Unmarshal(body, &unJson); errJU != nil {
		return nil, &apiBadHttpBody{body, errJU.Error()}
	}

	tasks = map[PkgMgrTask]struct{}{}

	if rootArray, rootIsArray := unJson.([]interface{}); rootIsArray {
		for _, task := range rootArray {
			if taskObject, taskIsObject := task.(map[string]interface{}); taskIsObject {
				nextTask := PkgMgrTask{}

				if packageName, hasPackageName := taskObject["package"]; hasPackageName {
					packageNameString, packageNameIsString := packageName.(string)
					if packageNameIsString && packageNameString != "" {
						nextTask.PackageName = packageNameString
					} else {
						return nil, &apiBadHttpBody{body, "package must be a non-empty string"}
					}
				} else {
					return nil, &apiBadHttpBody{body, "package missing"}
				}

				if action, hasAction := taskObject["action"]; hasAction {
					if actionString, actionIsString := action.(string); actionIsString && actionString != "" {
						if validAction, actionIsValid := api2pkgMgrActions[actionString]; actionIsValid {
							nextTask.Action = validAction
						} else {
							return nil, &apiBadHttpBody{body, fmt.Sprintf("bad action: %#v", actionString)}
						}
					} else {
						return nil, &apiBadHttpBody{body, "action must be a non-empty string"}
					}
				} else {
					return nil, &apiBadHttpBody{body, "action missing"}
				}

				if fromVersion, hasFromVersion := taskObject["from_version"]; hasFromVersion {
					fromVersionString, fromVersionIsString := fromVersion.(string)
					if fromVersionIsString && fromVersionString != "" {
						nextTask.FromVersion = fromVersionString
					} else {
						return nil, &apiBadHttpBody{body, "from_version must be a non-empty string"}
					}
				}

				if toVersion, hasToVersion := taskObject["to_version"]; hasToVersion {
					toVersionString, toVersionIsString := toVersion.(string)
					if toVersionIsString && toVersionString != "" {
						nextTask.ToVersion = toVersionString
					} else {
						return nil, &apiBadHttpBody{body, "to_version must be a non-empty string"}
					}
				}

				var hasVersions bool

				switch nextTask.Action {
				case PkgMgrInstall:
					hasVersions = nextTask.FromVersion == "" && nextTask.ToVersion != ""
				case PkgMgrUpdate:
					hasVersions = nextTask.FromVersion != "" && nextTask.ToVersion != ""
				case PkgMgrConfigure:
					hasVersions = (nextTask.FromVersion == "") != (nextTask.ToVersion == "")
				case PkgMgrRemove, PkgMgrPurge:
					hasVersions = nextTask.FromVersion != ""
				}

				if !hasVersions {
					return nil, &apiBadHttpBody{
						body, "too many/few versions for an " + taskObject["action"].(string) + " action",
					}
				}

				tasks[nextTask] = struct{}{}
			} else {
				return nil, &apiBadHttpBody{body, "task must be an object"}
			}
		}
	} else {
		return nil, &apiBadHttpBody{body, "must be an array"}
	}

	return
}
