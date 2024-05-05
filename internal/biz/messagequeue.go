package biz

type TransactionMessage struct {
	Id        string `json:"id"`
	UserId    string `json:"user_id"`
	Amount    string `json:"amount"`
	Symbol    string `json:"symbol"`
	TransType string `json:"type"`
}

type TransactionPublisher interface {
	Publish(msg *TransactionMessage)
}
