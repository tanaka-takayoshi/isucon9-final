package mock

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/util"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/jarcoal/httpmock"
)

// Mock は `isutrain` のモック実装です
type Mock struct {
	LoginDelay             time.Duration
	ListStationsDelay      time.Duration
	SearchTrainsDelay      time.Duration
	ListTrainSeatsDelay    time.Duration
	ReserveDelay           time.Duration
	CommitReservationDelay time.Duration
	CancelReservationDelay time.Duration
	ListReservationDelay   time.Duration

	injectFunc func(path string) error

	paymentMock *paymentMock
}

func NewMock(paymentMock *paymentMock) *Mock {
	return &Mock{
		injectFunc: func(path string) error {
			return nil
		},
		paymentMock: paymentMock,
	}
}

func (m *Mock) Inject(f func(path string) error) {
	m.injectFunc = f
}

func (m *Mock) Initialize(req *http.Request) ([]byte, int) {
	if err := m.injectFunc(req.URL.Path); err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}
	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// Register はユーザ登録を行います
func (m *Mock) Register(req *http.Request) ([]byte, int) {
	if err := req.ParseForm(); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	var (
		username = req.Form.Get("username")
		password = req.Form.Get("password")
		// TODO: 他にも登録情報を追加
	)
	if len(username) == 0 || len(password) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// Login はログイン処理結果を返します
func (m *Mock) Login(req *http.Request) ([]byte, int) {
	<-time.After(m.LoginDelay)
	if err := req.ParseForm(); err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	var (
		username = req.Form.Get("username")
		password = req.Form.Get("password")
	)
	if len(username) == 0 || len(password) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

func (m *Mock) ListStations(req *http.Request) ([]byte, int) {
	<-time.After(m.ListStationsDelay)
	b, err := json.Marshal([]*isutrain.Station{
		&isutrain.Station{ID: 1, Name: "isutrain1", IsStopExpress: false, IsStopSemiExpress: false, IsStopLocal: false},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// SearchTrains は新幹線検索結果を返します
func (m *Mock) SearchTrains(req *http.Request) ([]byte, int) {
	<-time.After(m.SearchTrainsDelay)
	query := req.URL.Query()
	if query == nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	// TODO: 検索クエリを受け取る
	// いつ(use_at)、どこからどこまで(from, to), 人数(number of people) で結果が帰って来れば良い
	// 日付を投げてきていて、DB称号可能などこからどこまでがあればいい
	// どこからどこまでは駅名を書く(IDはユーザから見たらまだわからない)
	useAt, err := util.ParseISO8601(query.Get("use_at"))
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}
	if useAt.IsZero() {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	var (
		from = query.Get("from")
		to   = query.Get("to")
	)
	if len(from) == 0 || len(to) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	b, err := json.Marshal(&isutrain.Trains{
		&isutrain.Train{Class: "のぞみ", Name: "96号", Start: 1, Last: 2},
		&isutrain.Train{Class: "こだま", Name: "96号", Start: 3, Last: 4},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// ListTrainSeats は列車の席一覧を返します
func (m *Mock) ListTrainSeats(req *http.Request) ([]byte, int) {
	<-time.After(m.ListTrainSeatsDelay)
	q := req.URL.Query()
	if q == nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	// 列車特定情報を受け取る
	var (
		trainClass = q["train_class"]
		trainName  = q["train_name"]
	)
	if len(trainClass) == 0 || len(trainName) == 0 {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	// 適当な席を返す
	b, err := json.Marshal(&isutrain.TrainSeats{
		&isutrain.TrainSeat{},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	return b, http.StatusOK
}

// Reserve は座席予約を実施し、結果を返します
func (m *Mock) Reserve(req *http.Request) ([]byte, int) {
	<-time.After(m.ReserveDelay)
	// 予約情報を受け取って、予約できたかを返す

	// FIXME: ユーザID
	// 複数の座席指定で予約するかもしれない
	// なので、予約には複数の座席予約が紐づいている

	b, err := json.Marshal(&isutrain.ReservationResponse{
		ReservationID: "1111111111",
		IsOk:          true,
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
	}

	// NOTE: とりあえず、パラメータガン無視でPOSTできるところ先にやる
	return b, http.StatusAccepted
}

// CommitReservation は予約を確定します
func (m *Mock) CommitReservation(req *http.Request) ([]byte, int) {
	<-time.After(m.CommitReservationDelay)
	// 予約IDを受け取って、確定するだけ

	_, err := httpmock.GetSubmatchAsUint(req, 1)
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	// FIXME: ちゃんとした決済情報を追加する
	m.paymentMock.addPaymentInformation()

	return []byte(http.StatusText(http.StatusAccepted)), http.StatusAccepted
}

// CancelReservation は予約をキャンセルします
func (m *Mock) CancelReservation(req *http.Request) ([]byte, int) {
	<-time.After(m.CancelReservationDelay)
	// 予約IDを受け取って

	_, err := httpmock.GetSubmatchAsUint(req, 1)
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return []byte(http.StatusText(http.StatusNoContent)), http.StatusNoContent
}

// ListReservations はアカウントにひもづく予約履歴を返します
func (m *Mock) ListReservations(req *http.Request) ([]byte, int) {
	<-time.After(m.ListReservationDelay)
	b, err := json.Marshal(isutrain.SeatReservations{
		&isutrain.SeatReservation{ID: 1111, PaymentMethod: string(isutrain.CreditCard), Status: string(isutrain.Pending), ReserveAt: time.Now()},
	})
	if err != nil {
		return []byte(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
	}

	return b, http.StatusOK
}