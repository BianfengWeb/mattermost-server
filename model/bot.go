// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"
)

const (
	BOT_DISPLAY_NAME_MAX_RUNES = USER_FIRST_NAME_MAX_RUNES
)

// Bot is a special type of User meant for programmatic interactions.
// Note that the primary key of a bot is the UserId, and matches the primary key of the
// corresponding user.
type Bot struct {
	UserId      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	CreatorId   string `json:"creator_id"`
	CreateAt    int64  `json:"create_at"`
	UpdateAt    int64  `json:"update_at"`
	DeleteAt    int64  `json:"delete_at"`
}

// BotPatch is a description of what fields to update on an existing bot.
type BotPatch struct {
	Username    *string `json:"username"`
	DisplayName *string `json:"display_name"`
	Description *string `json:"description"`
}

// BotList is a list of bots.
type BotList []*Bot

// Trace describes the minimum information required to identify a bot for the purpose of logging.
func (b *Bot) Trace() map[string]interface{} {
	return map[string]interface{}{"user_id": b.UserId}
}

// Clone returns a shallow copy of the bot.
func (b *Bot) Clone() *Bot {
	copy := *b
	return &copy
}

// IsValid validates the bot and returns an error if it isn't configured correctly.
func (b *Bot) IsValid() *AppError {
	if len(b.UserId) != 26 {
		return NewAppError("Bot.IsValid", "model.bot.is_valid.user_id.app_error", b.Trace(), "", http.StatusBadRequest)
	}

	if !IsValidUsername(b.Username) {
		return NewAppError("Bot.IsValid", "model.bot.is_valid.username.app_error", b.Trace(), "", http.StatusBadRequest)
	}

	if utf8.RuneCountInString(b.DisplayName) > BOT_DISPLAY_NAME_MAX_RUNES {
		return NewAppError("Bot.IsValid", "model.bot.is_valid.user_id.app_error", b.Trace(), "", http.StatusBadRequest)
	}

	if utf8.RuneCountInString(b.Description) > 1024 {
		return NewAppError("Bot.IsValid", "model.bot.is_valid.description.app_error", b.Trace(), "", http.StatusBadRequest)
	}

	if len(b.CreatorId) != 26 {
		return NewAppError("Bot.IsValid", "model.bot.is_valid.creator_id.app_error", b.Trace(), "", http.StatusBadRequest)
	}

	if b.CreateAt == 0 {
		return NewAppError("Bot.IsValid", "model.bot.is_valid.create_at.app_error", b.Trace(), "", http.StatusBadRequest)
	}

	if b.UpdateAt == 0 {
		return NewAppError("Bot.IsValid", "model.bot.is_valid.update_at.app_error", b.Trace(), "", http.StatusBadRequest)
	}

	return nil
}

// PreSave should be run before saving a new bot to the database.
func (b *Bot) PreSave() {
	b.CreateAt = GetMillis()
	b.UpdateAt = b.CreateAt
	b.DeleteAt = 0
}

// PreUpdate should be run before saving an updated bot to the database.
func (b *Bot) PreUpdate() {
	b.UpdateAt = GetMillis()
}

// Etag generates an etag for caching.
func (b *Bot) Etag() string {
	return Etag(b.UserId, b.UpdateAt)
}

// ToJson serializes the bot to json.
func (b *Bot) ToJson() []byte {
	data, _ := json.Marshal(b)
	return data
}

// BotFromJson deserializes a bot from json.
func BotFromJson(data io.Reader) *Bot {
	var bot *Bot
	json.NewDecoder(data).Decode(&bot)
	return bot
}

// Patch modifies an existing bot with optional fields from the given patch.
func (b *Bot) Patch(patch *BotPatch) {
	if patch.Username != nil {
		b.Username = *patch.Username
	}

	if patch.DisplayName != nil {
		b.DisplayName = *patch.DisplayName
	}

	if patch.Description != nil {
		b.Description = *patch.Description
	}
}

// ToJson serializes the bot patch to json.
func (b *BotPatch) ToJson() []byte {
	data, err := json.Marshal(b)
	if err != nil {
		return nil
	}

	return data
}

// BotPatchFromJson deserializes a bot patch from json.
func BotPatchFromJson(data io.Reader) *BotPatch {
	decoder := json.NewDecoder(data)
	var botPatch BotPatch
	err := decoder.Decode(&botPatch)
	if err != nil {
		return nil
	}

	return &botPatch
}

// UserFromBotModel returns a user model describing the bot fields stored in the User store.
func UserFromBotModel(b *Bot) *User {
	return &User{
		Id:       b.UserId,
		Username: b.Username,
		// TODO: Allow users not to have an email.
		Email:     fmt.Sprintf("%s@localhost", strings.ToLower(b.Username)),
		FirstName: b.DisplayName,
	}
}

// BotListFromJson deserializes a list of bots from json.
func BotListFromJson(data io.Reader) BotList {
	var bots BotList
	json.NewDecoder(data).Decode(&bots)
	return bots
}

// ToJson serializes a list of bots to json.
func (l *BotList) ToJson() []byte {
	b, _ := json.Marshal(l)
	return b
}

// Etag computes the etag for a list of bots.
func (l *BotList) Etag() string {
	id := "0"
	var t int64 = 0
	var delta int64 = 0

	for _, v := range *l {
		if v.UpdateAt > t {
			t = v.UpdateAt
			id = v.UserId
		}

	}

	return Etag(id, t, delta, len(*l))
}
