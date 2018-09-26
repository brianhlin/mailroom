package flows

import (
	"time"

	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/utils"
)

// NodeUUID is a UUID of a flow node
type NodeUUID utils.UUID

// ExitUUID is the UUID of a node exit
type ExitUUID utils.UUID

// ActionUUID is the UUID of an action
type ActionUUID utils.UUID

// ContactID is the ID of a contact
type ContactID int64

// ContactUUID is the UUID of a contact
type ContactUUID utils.UUID

// RunUUID is the UUID of a flow run
type RunUUID utils.UUID

// StepUUID is the UUID of a run step
type StepUUID utils.UUID

// InputUUID is the UUID of an input
type InputUUID utils.UUID

// MsgID is the ID of a message
type MsgID int64

// NilMsgID is our constant for nil message ids
const NilMsgID = MsgID(0)

// MsgUUID is the UUID of a message
type MsgUUID utils.UUID

// FlowType represents the different types of flows
type FlowType string

// UINodeType tells the editor how to render a particular node
type UINodeType string

// UINodeConfig contains config unique to its type
type UINodeConfig map[string]interface{}

// Sticky is a migrated note
type Sticky map[string]interface{}

const (
	// FlowTypeMessaging is a flow that is run over a messaging channel
	FlowTypeMessaging FlowType = "messaging"

	// FlowTypeMessagingOffline is a flow which is run over an offline messaging client like Surveyor
	FlowTypeMessagingOffline FlowType = "messaging_offline"

	// FlowTypeVoice is a flow which is run over IVR
	FlowTypeVoice FlowType = "voice"
)

// SessionStatus represents the current status of the engine session
type SessionStatus string

const (
	// SessionStatusActive represents a session that is still active
	SessionStatusActive SessionStatus = "active"

	// SessionStatusCompleted represents a session that has run to completion
	SessionStatusCompleted SessionStatus = "completed"

	// SessionStatusWaiting represents a session which is waiting for something from the caller
	SessionStatusWaiting SessionStatus = "waiting"

	// SessionStatusErrored represents a session that encountered an error
	SessionStatusErrored SessionStatus = "errored"
)

// RunStatus represents the current status of the flow run
type RunStatus string

const (
	// RunStatusActive represents a run that is still active
	RunStatusActive RunStatus = "active"

	// RunStatusCompleted represents a run that has run to completion
	RunStatusCompleted RunStatus = "completed"

	// RunStatusWaiting represents a run which is waiting for something from the caller
	RunStatusWaiting RunStatus = "waiting"

	// RunStatusErrored represents a run that encountered an error
	RunStatusErrored RunStatus = "errored"

	// RunStatusExpired represents a run that expired due to inactivity
	RunStatusExpired RunStatus = "expired"

	// RunStatusInterrupted represents a run that was interrupted by another flow
	RunStatusInterrupted RunStatus = "interrupted"
)

type FlowAssets interface {
	Get(assets.FlowUUID) (Flow, error)
}

// SessionAssets is the assets available to a session
type SessionAssets interface {
	Channels() *ChannelAssets
	Fields() *FieldAssets
	Flows() FlowAssets
	Groups() *GroupAssets
	Labels() *LabelAssets
	Locations() *LocationAssets
	Resthooks() *ResthookAssets
}

// Flow describes the ordered logic of actions and routers. It renders as its name in a template, and has the following
// properties which can be accessed:
//
//  * `uuid` the UUID of the flow
//  * `name` the name of the flow
//  * `revision` the revision number of the flow
//
// Examples:
//
//   @run.flow -> Registration
//   @child.flow -> Collect Age
//   @run.flow.uuid -> 50c3706e-fedb-42c0-8eab-dda3335714b7
//   @(json(run.flow)) -> {"name":"Registration","revision":123,"uuid":"50c3706e-fedb-42c0-8eab-dda3335714b7"}
//
// @context flow
type Flow interface {
	types.XValue
	types.XResolvable

	UUID() assets.FlowUUID
	Name() string
	Revision() int
	Language() utils.Language
	ExpireAfterMinutes() int
	Localization() Localization
	UI() UI

	Validate(SessionAssets) error
	Nodes() []Node
	GetNode(uuid NodeUUID) Node

	Reference() *assets.FlowReference
}

// Node is a single node in a flow
type Node interface {
	UUID() NodeUUID

	Actions() []Action
	AddAction(Action)

	Router() Router
	Exits() []Exit
	Wait() Wait
}

// Action is an action within a flow node
type Action interface {
	UUID() ActionUUID

	Execute(FlowRun, Step, EventLog) error
	Validate(SessionAssets) error
	AllowedFlowTypes() []FlowType
	utils.Typed
}

type Router interface {
	PickRoute(FlowRun, []Exit, Step) (*string, Route, error)
	Validate([]Exit) error
	ResultName() string
	utils.Typed
}

type Exit interface {
	UUID() ExitUUID
	DestinationNodeUUID() NodeUUID
	Name() string
}

type Wait interface {
	utils.Typed

	Timeout() *int
	TimeoutOn() *time.Time

	Begin(FlowRun, Step)
	CanResume([]CallerEvent) bool
}

// Localization provide a way to get the translations for a specific language
type Localization interface {
	AddItemTranslation(utils.Language, utils.UUID, string, []string)
	GetTranslations(utils.Language) Translations
	Languages() utils.LanguageList
}

// Translations provide a way to get the translation for a specific language for a uuid/key pair
type Translations interface {
	GetTextArray(uuid utils.UUID, key string) []string
}

// UINodeDetails is the top level ui details for a node
type UINodeDetails interface {
}

// UI is a optional section in a flow definition with editor specific information
type UI interface {
	AddNode(uuid NodeUUID, details UINodeDetails)
	AddSticky(sticky Sticky)

	GetNode(uuid NodeUUID) UINodeDetails
}

// Position holds coordinates for a node
type Position interface {
}

// Trigger represents something which can initiate a session with the flow engine. It has several properties which can be
// accessed in expressions:
//
//  * `type` the type of the trigger, one of "manual" or "flow"
//  * `params` the parameters passed to the trigger
//
// Examples:
//
//   @trigger.type -> flow_action
//   @trigger.params -> {"source": "website","address": {"state": "WA"}}
//   @(json(trigger)) -> {"params":{"source":"website","address":{"state":"WA"}},"type":"flow_action"}
//
// @context trigger
type Trigger interface {
	utils.Typed
	types.XValue
	types.XResolvable

	Environment() utils.Environment
	Flow() *assets.FlowReference
	Contact() *Contact
	Params() types.XValue
	TriggeredOn() time.Time
}

// EventOrigin is the allowed origin of an event
type EventOrigin int

const (
	// EventOriginCaller means an event can originate from the caller
	EventOriginCaller EventOrigin = 1

	// EventOriginEngine means an event can originate from the engine
	EventOriginEngine EventOrigin = 2
)

// Event describes a state change
type Event interface {
	utils.Typed

	CreatedOn() time.Time
	StepUUID() StepUUID
	SetStepUUID(StepUUID)
}

// CallerEvent is an event we can receive from our caller
type CallerEvent interface {
	Event

	Validate(SessionAssets) error
	Apply(FlowRun) error
}

// EventLog is the log of events the caller must apply after each call
type EventLog interface {
	Add(Event)
	Events() []Event
}

// Input describes input from the contact and currently we only support one type of input: `msg`. Any input has the following
// properties which can be accessed:
//
//  * `uuid` the UUID of the input
//  * `type` the type of the input, e.g. `msg`
//  * `channel` the [channel](#context:channel) that the input was received on
//  * `created_on` the time when the input was created
//
// An input of type `msg` renders as its text and attachments in a template, and has the following additional properties:
//
//  * `text` the text of the message
//  * `attachments` any [attachments](#context:attachment) on the message
//  * `urn` the [URN](#context:urn) that the input was received on
//
// Examples:
//
//   @input -> Hi there\nhttp://s3.amazon.com/bucket/test.jpg\nhttp://s3.amazon.com/bucket/test.mp3
//   @input.type -> msg
//   @input.text -> Hi there
//   @input.attachments -> ["http://s3.amazon.com/bucket/test.jpg","http://s3.amazon.com/bucket/test.mp3"]
//   @(json(run.input)) -> {"attachments":[{"content_type":"image/jpeg","url":"http://s3.amazon.com/bucket/test.jpg"},{"content_type":"audio/mp3","url":"http://s3.amazon.com/bucket/test.mp3"}],"channel":{"address":"+12345671111","name":"My Android Phone","uuid":"57f1078f-88aa-46f4-a59a-948a5739c03d"},"created_on":"2000-01-01T00:00:00.000000Z","text":"Hi there","type":"msg","urn":{"display":"","path":"+12065551212","scheme":"tel"},"uuid":"9bf91c2b-ce58-4cef-aacc-281e03f69ab5"}
//
// @context input
type Input interface {
	types.XValue
	utils.Typed

	UUID() InputUUID
	CreatedOn() time.Time
	Channel() *Channel
}

type Step interface {
	types.XValue
	types.XResolvable

	UUID() StepUUID
	NodeUUID() NodeUUID
	ExitUUID() ExitUUID
	ArrivedOn() time.Time

	Leave(ExitUUID)
}

type EngineConfig interface {
	DisableWebhooks() bool
	WebhookMocks() []*WebhookMock
	MaxWebhookResponseBytes() int
}

// Session represents the session of a flow run which may contain many runs
type Session interface {
	Assets() SessionAssets

	Environment() utils.Environment
	SetEnvironment(utils.Environment)

	Contact() *Contact
	SetContact(*Contact)

	Status() SessionStatus
	Trigger() Trigger
	PushFlow(Flow, FlowRun)
	Wait() Wait
	FlowOnStack(assets.FlowUUID) bool

	Start(Trigger, []CallerEvent) error
	Resume([]CallerEvent) error
	Runs() []FlowRun
	GetRun(RunUUID) (FlowRun, error)
	GetCurrentChild(FlowRun) FlowRun
	ParentRun() RunSummary

	Events() []Event
	LogEvent(Event)

	EngineConfig() EngineConfig
	HTTPClient() *utils.HTTPClient
}

// RunSummary represents the minimum information available about all runs (current or related) and is the
// representation of runs made accessible to router tests.
type RunSummary interface {
	UUID() RunUUID
	Contact() *Contact
	Flow() Flow
	Status() RunStatus
	Results() Results
}

// RunEnvironment is a run specific environment which adds location functionality required by some router tests
type RunEnvironment interface {
	utils.Environment

	FindLocations(string, utils.LocationLevel, *utils.Location) ([]*utils.Location, error)
	FindLocationsFuzzy(string, utils.LocationLevel, *utils.Location) ([]*utils.Location, error)
	LookupLocation(LocationPath) (*utils.Location, error)
}

// FlowRun is a single contact's journey through a flow. It records the path they have taken, and the results that have been
// collected. It has several properties which can be accessed in expressions:
//
//  * `uuid` the UUID of the run
//  * `flow` the [flow](#context:flow) of the run
//  * `contact` the [contact](#context:contact) of the flow run
//  * `input` the [input](#context:input) of the current run
//  * `results` the results that have been saved for this run
//  * `results.[snaked_result_name]` the value of the specific result, e.g. `results.age`
//
// Examples:
//
//   @run.flow.name -> Registration
//
// @context run
type FlowRun interface {
	types.XValue
	types.XResolvable
	RunSummary

	Environment() RunEnvironment
	Session() Session
	Context() types.XValue
	Input() Input

	SetContact(*Contact)
	SetInput(Input)
	SetStatus(RunStatus)

	AddEvent(Step, Action, Event)
	AddError(Step, Action, error)
	AddFatalError(Step, Action, error)

	CreateStep(Node) Step
	Path() []Step
	PathLocation() (Step, Node, error)
	Events() []Event

	EvaluateTemplate(template string) (types.XValue, error)
	EvaluateTemplateAsString(template string, urlEncode bool) (string, error)

	GetText(utils.UUID, string, string) string
	GetTextArray(utils.UUID, string, []string) []string
	GetTranslatedTextArray(utils.UUID, string, []string, utils.LanguageList) []string

	Snapshot() RunSummary
	Parent() RunSummary
	ParentInSession() FlowRun
	Ancestors() []FlowRun

	CreatedOn() time.Time
	ExpiresOn() *time.Time
	ResetExpiration(*time.Time)
	ExitedOn() *time.Time
	Exit(RunStatus)
}
