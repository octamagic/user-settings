# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [common_session.proto](#common_session-proto)
    - [Geo](#proto-soundcloud-common-session-Geo)
    - [RequestedFeature](#proto-soundcloud-common-session-RequestedFeature)
    - [UserSession](#proto-soundcloud-common-session-UserSession)
  
- [user_settings_api.proto](#user_settings_api-proto)
    - [DeleteSettingsRequest](#proto-soundcloud-user_settings-api-DeleteSettingsRequest)
    - [DeleteSettingsResponse](#proto-soundcloud-user_settings-api-DeleteSettingsResponse)
    - [GetBatchSettingsRequest](#proto-soundcloud-user_settings-api-GetBatchSettingsRequest)
    - [GetBatchSettingsResponse](#proto-soundcloud-user_settings-api-GetBatchSettingsResponse)
    - [GetSettingsRequest](#proto-soundcloud-user_settings-api-GetSettingsRequest)
    - [GetSettingsResponse](#proto-soundcloud-user_settings-api-GetSettingsResponse)
    - [PutSettingsRequest](#proto-soundcloud-user_settings-api-PutSettingsRequest)
    - [PutSettingsResponse](#proto-soundcloud-user_settings-api-PutSettingsResponse)
    - [Setting](#proto-soundcloud-user_settings-api-Setting)
    - [Setting.ContextEntry](#proto-soundcloud-user_settings-api-Setting-ContextEntry)
    - [UserSettings](#proto-soundcloud-user_settings-api-UserSettings)
  
    - [SettingType](#proto-soundcloud-user_settings-api-SettingType)
  
    - [UserSettingsService](#proto-soundcloud-user_settings-api-UserSettingsService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="common_session-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## common_session.proto



<a name="proto-soundcloud-common-session-Geo"></a>

### Geo
the geographic location, determined from IP geolocation.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| country_code | [string](#string) |  | two-letter country code, or &#39;--&#39; if unknown |
| city | [google.protobuf.StringValue](#google-protobuf-StringValue) |  |  |
| region | [google.protobuf.StringValue](#google-protobuf-StringValue) |  |  |






<a name="proto-soundcloud-common-session-RequestedFeature"></a>

### RequestedFeature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="proto-soundcloud-common-session-UserSession"></a>

### UserSession
Metadata about the interaction between a user and the SoundCloud systems


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_urn | [google.protobuf.StringValue](#google-protobuf-StringValue) |  | user making the request |
| agent_urn | [google.protobuf.StringValue](#google-protobuf-StringValue) |  | the SoundCloud property being used for this session |
| geo | [Geo](#proto-soundcloud-common-session-Geo) |  | the geographic location of the user |
| scopes | [string](#string) | repeated | OAuth scopes for this session |
| features | [string](#string) | repeated | business features that are enabled for this session |
| app_variant_ids | [int32](#int32) | repeated | propagate variants for github.com/soundcloud/ab-testing |
| app_requested_features | [RequestedFeature](#proto-soundcloud-common-session-RequestedFeature) | repeated | features requested in Firebase A/B tests |





 

 

 

 



<a name="user_settings_api-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## user_settings_api.proto



<a name="proto-soundcloud-user_settings-api-DeleteSettingsRequest"></a>

### DeleteSettingsRequest
Delete setting


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_session | [proto.soundcloud.common.session.UserSession](#proto-soundcloud-common-session-UserSession) |  |  |
| user_urn | [string](#string) |  |  |
| setting_id | [string](#string) |  |  |






<a name="proto-soundcloud-user_settings-api-DeleteSettingsResponse"></a>

### DeleteSettingsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | [bool](#bool) |  |  |






<a name="proto-soundcloud-user_settings-api-GetBatchSettingsRequest"></a>

### GetBatchSettingsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_session | [proto.soundcloud.common.session.UserSession](#proto-soundcloud-common-session-UserSession) |  |  |
| user_urns | [string](#string) | repeated |  |
| setting_id | [string](#string) |  |  |






<a name="proto-soundcloud-user_settings-api-GetBatchSettingsResponse"></a>

### GetBatchSettingsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| settings | [UserSettings](#proto-soundcloud-user_settings-api-UserSettings) | repeated |  |






<a name="proto-soundcloud-user_settings-api-GetSettingsRequest"></a>

### GetSettingsRequest
Get settings request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_session | [proto.soundcloud.common.session.UserSession](#proto-soundcloud-common-session-UserSession) |  |  |
| user_urn | [string](#string) |  |  |
| types | [SettingType](#proto-soundcloud-user_settings-api-SettingType) | repeated | Filter by type |
| setting_ids | [string](#string) | repeated | Filter by ID user_urn &#43; setting_catigory |






<a name="proto-soundcloud-user_settings-api-GetSettingsResponse"></a>

### GetSettingsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| settings | [UserSettings](#proto-soundcloud-user_settings-api-UserSettings) |  |  |






<a name="proto-soundcloud-user_settings-api-PutSettingsRequest"></a>

### PutSettingsRequest
Create/update setting


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_session | [proto.soundcloud.common.session.UserSession](#proto-soundcloud-common-session-UserSession) |  |  |
| setting | [Setting](#proto-soundcloud-user_settings-api-Setting) |  |  |






<a name="proto-soundcloud-user_settings-api-PutSettingsResponse"></a>

### PutSettingsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| setting | [Setting](#proto-soundcloud-user_settings-api-Setting) |  |  |






<a name="proto-soundcloud-user_settings-api-Setting"></a>

### Setting
Generic setting struct


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| user_urn | [string](#string) |  |  |
| type | [SettingType](#proto-soundcloud-user_settings-api-SettingType) |  |  |
| updated_at | [int64](#int64) |  |  |
| schema_version | [int32](#int32) |  |  |
| value | [string](#string) |  | Use a single field to store the value as a JSON string. |
| context | [Setting.ContextEntry](#proto-soundcloud-user_settings-api-Setting-ContextEntry) | repeated |  |






<a name="proto-soundcloud-user_settings-api-Setting-ContextEntry"></a>

### Setting.ContextEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="proto-soundcloud-user_settings-api-UserSettings"></a>

### UserSettings
Settings collection


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_urn | [string](#string) |  |  |
| settings | [Setting](#proto-soundcloud-user_settings-api-Setting) | repeated |  |





 


<a name="proto-soundcloud-user_settings-api-SettingType"></a>

### SettingType
Setting types enum

| Name | Number | Description |
| ---- | ------ | ----------- |
| SETTING_TYPE_UNSPECIFIED | 0 |  |
| SETTING_TYPE_BADGE | 1 |  |
| SETTING_TYPE_FEATURED_FAN | 2 |  |
| SETTING_TYPE_NOTIFICATION | 3 |  |
| SETTING_TYPE_PRIVACY | 4 |  |
| SETTING_TYPE_ANALYTICS | 5 |  |
| SETTING_TYPE_FEATURED_PROFILE | 6 |  |
| SETTING_TYPE_THEME | 7 |  |
| SETTING_TYPE_LANGUAGE | 8 |  |
| SETTING_TYPE_PROTECTION_ENABLED | 9 | MINOR PROTECTION SETTING |


 

 


<a name="proto-soundcloud-user_settings-api-UserSettingsService"></a>

### UserSettingsService
Main service CRUD

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetSettings | [GetSettingsRequest](#proto-soundcloud-user_settings-api-GetSettingsRequest) | [GetSettingsResponse](#proto-soundcloud-user_settings-api-GetSettingsResponse) |  |
| GetBatchSettings | [GetBatchSettingsRequest](#proto-soundcloud-user_settings-api-GetBatchSettingsRequest) | [GetBatchSettingsResponse](#proto-soundcloud-user_settings-api-GetBatchSettingsResponse) |  |
| PutSettings | [PutSettingsRequest](#proto-soundcloud-user_settings-api-PutSettingsRequest) | [PutSettingsResponse](#proto-soundcloud-user_settings-api-PutSettingsResponse) |  |
| DeleteSettings | [DeleteSettingsRequest](#proto-soundcloud-user_settings-api-DeleteSettingsRequest) | [DeleteSettingsResponse](#proto-soundcloud-user_settings-api-DeleteSettingsResponse) |  |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

