package main

import "time"

type YFRequestConversations struct {
	Quote     string `json:"quote" validate:"required"`
	StartTime int64    `json:"startTime" validate:"required"`
	EndTime   int64   `json:"endTime" validate:"required"`
}

type TwitterRecentSearch struct {
	Quote     string `json:"quote" validate:"required"`
	StartTime int64    `json:"startTime" validate:"required"`
	EndTime   int64   `json:"endTime" validate:"required"`
}

type RedditRecentSearch struct {
	Quote     string `json:"quote" validate:"required"`
	StartTime int64    `json:"startTime" validate:"required"`
	EndTime   int64   `json:"endTime" validate:"required"`
}

type TimeRange struct{
	StartTime int64    `json:"startTime" validate:"required"`
	EndTime   int64   `json:"endTime" validate:"required"`
}

type ConversationsResult struct {
	Filename string
	LineNum int
}

type TwitterRecentSearchResult struct {
	Data []struct {
		Text      string    `json:"text"`
		Lang      string    `json:"lang"`
		CreatedAt time.Time `json:"created_at"`
		ID        string    `json:"id"`
		AuthorID  string    `json:"author_id"`
	} `json:"data"`
	Meta struct {
		NewestID    string `json:"newest_id"`
		OldestID    string `json:"oldest_id"`
		ResultCount int    `json:"result_count"`
		NextToken   string `json:"next_token"`
	} `json:"meta"`
}