package games

type AnswerStats struct {
	Option     int64
	Votes      int64
	AnswerText string
}

type UserScore struct {
	Username string
	Score    int64
}

type GameResults struct {
	AnswersData    []AnswerStats
	ScoreboardData []UserScore
}

type Game struct {
	GameId           int64
	GameName         string
	GameTotalMembers int64
}
type GameInvite struct {
	GameName    string
	GameId      int64
	CreatorName string
	CreatorId   int64
}

type GamesPageData struct {
	CurrentGames []Game
	GameInvites  []GameInvite
}

type ResultsData struct {
	AnswersData    []AnswerStats
	ScoreboardData []UserScore
}

type GamePlayData struct {
	ShowResults int
	Data        interface{}
}

type QuestionData struct{}
