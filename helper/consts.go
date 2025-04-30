package helper

// 设备端类型
const (
	DeviceTypeWeb     = "24" //web
	DeviceTypeH5      = "25" //h5
	DeviceTypeAndroid = "26" //android
	DeviceTypeIOS     = "27" //ios
	DeviceTypePwa     = "28" //pwa
)

var (
	DeviceTypes = map[string]string{
		DeviceTypeWeb:     "1", //web
		DeviceTypeH5:      "2", //h5
		DeviceTypeAndroid: "3", //android
		DeviceTypeIOS:     "4", //ios
		DeviceTypePwa:     "5", //pwa
	}
)
