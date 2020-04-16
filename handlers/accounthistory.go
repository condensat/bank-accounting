// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"
	"time"

	"github.com/condensat/bank-accounting/common"
	"github.com/condensat/bank-accounting/internal"
	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func AccountHistory(ctx context.Context, accountID uint64, from, to time.Time) (string, []common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountHistory")

	log = log.WithFields(logrus.Fields{
		"AccountID": accountID,
		"From":      from,
		"To":        to,
	})

	// Database Query
	db := appcontext.Database(ctx)
	account, err := database.GetAccountByID(db, model.AccountID(accountID))
	if err != nil {
		return "", nil, err
	}

	operations, err := database.GeAccountHistoryRange(db, account.ID, from, to)
	if err != nil {
		return "", nil, err
	}

	var result []common.AccountEntry
	for _, op := range operations {
		if !op.IsValid() {
			log.WithError(database.ErrInvalidAccountOperation).
				Warn("Invalid operation in history")
			continue
		}

		result = append(result, common.AccountEntry{
			OperationID:     uint64(op.ID),
			OperationPrevID: uint64(op.PrevID),

			AccountID: uint64(op.AccountID),
			Currency:  string(account.CurrencyName),

			OperationType:    string(op.OperationType),
			SynchroneousType: string(op.SynchroneousType),

			Timestamp: op.Timestamp,
			Label:     "N/A",
			Amount:    float64(*op.Amount),
			Balance:   float64(*op.Balance),

			LockAmount:  float64(*op.LockAmount),
			TotalLocked: float64(*op.TotalLocked),
		})
	}

	log.
		WithField("Count", len(result)).
		Debug("Account history retrieved")

	return string(account.CurrencyName), result, nil
}

func OnAccountHistory(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountHistory")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountHistory
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"AccountID": request.AccountID,
			})

			currency, entries, err := AccountHistory(ctx, request.AccountID, request.From, request.To)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to get AccountHistory")
				return nil, internal.ErrInternalError
			}

			// create & return response
			return &common.AccountHistory{
				AccountID: request.AccountID,
				Currency:  currency,
				From:      request.From,
				To:        request.To,

				Entries: entries,
			}, nil
		})
}