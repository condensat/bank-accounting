// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-accounting/common"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query"

	"github.com/sirupsen/logrus"
)

func AccountList(ctx context.Context, userID uint64) ([]common.AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountList")
	var result []common.AccountInfo

	log = log.WithField("UserID", userID)

	// Acquire Lock
	lock, err := cache.LockUser(ctx, userID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock user")
		return result, cache.ErrLockError
	}
	defer lock.Unlock()

	// Database Query
	db := appcontext.Database(ctx)
	err = db.Transaction(func(db database.Context) error {
		accounts, err := query.GetAccountsByUserAndCurrencyAndName(db, model.UserID(userID), "*", "*")
		if err != nil {
			return err
		}

		for _, account := range accounts {

			account, err := txGetAccountInfo(db, account)
			if err != nil {
				return err
			}

			result = append(result, account)
		}

		return nil
	})

	if err == nil {
		log.WithField("Count", len(result)).
			Debug("User accounts retrieved")
	}

	return result, err
}

func OnAccountList(ctx context.Context, subject string, message *messaging.Message) (*messaging.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountList")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.UserAccounts
	return messaging.HandleRequest(ctx, appcontext.AppName(ctx), message, &request,
		func(ctx context.Context, _ messaging.BankObject) (messaging.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"UserID": request.UserID,
			})

			accounts, err := AccountList(ctx, request.UserID)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to list user accounts")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &common.UserAccounts{
				UserID:   request.UserID,
				Accounts: accounts[:],
			}, nil
		})
}
