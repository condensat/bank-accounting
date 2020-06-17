// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-accounting/common"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/sirupsen/logrus"
)

const (
	BankWitdrawAccountName = model.AccountName("withdraw")
)

func AccountTransferWithdraw(ctx context.Context, withdraw common.AccountTransferWithdraw) (common.AccountTransfer, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountTransferWithdraw")

	bankAccountID, err := getBankWithdrawAccount(ctx, withdraw.Source.Currency)
	if err != nil {
		log.WithError(err).
			Error("Invalid BankAccount")
		return common.AccountTransfer{}, database.ErrInvalidAccountID
	}

	log = log.WithFields(logrus.Fields{
		"BankAccountId": bankAccountID,
		"Currency":      withdraw.Source.Currency,
	})

	amount := withdraw.Source.Amount

	if amount <= 0.0 {
		return common.AccountTransfer{}, database.ErrInvalidWithdrawAmount
	}

	batchMode := model.BatchModeNormal
	if len(withdraw.BatchMode) > 0 {
		batchMode = model.BatchMode(withdraw.BatchMode)
	}

	var referenceID uint64
	// Database Query
	db := appcontext.Database(ctx)
	err = db.Transaction(func(db bank.Database) error {
		w, err := database.AddWithdraw(db,
			model.AccountID(withdraw.Source.AccountID),
			model.AccountID(bankAccountID),
			model.Float(amount), batchMode,
			"{}",
		)
		if err != nil {
			log.WithError(err).
				Error("AddWithdraw failed")
			return err
		}
		_, err = database.AddWithdrawInfo(db, w.ID, model.WithdrawStatusCreated, "{}")
		if err != nil {
			log.WithError(err).
				Error("AddWithdrawInfo failed")
			return err
		}

		referenceID = uint64(w.ID)

		return nil
	})

	if err != nil {
		return common.AccountTransfer{}, err
	}

	result, err := AccountTransfer(ctx, common.AccountTransfer{
		Source: withdraw.Source,
		Destination: common.AccountEntry{
			AccountID: uint64(bankAccountID),

			OperationType:    withdraw.Source.OperationType,
			SynchroneousType: "async-start",
			ReferenceID:      referenceID,

			Timestamp: time.Now(),
			Amount:    amount,

			Label: withdraw.Source.Label,

			LockAmount: amount,
			Currency:   withdraw.Source.Currency,
		},
	})

	if err == nil {
		log.Debug("AccountWithdraw created")
	}

	return result, err
}

func OnAccountTransferWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountTransferWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountTransferWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			response, err := AccountTransferWithdraw(ctx, request)
			if err != nil {
				log.WithError(err).
					WithFields(logrus.Fields{
						"AccountID": request.Source.AccountID,
					}).Errorf("Failed to AccountTransferWithdraw")
				return nil, cache.ErrInternalError
			}

			// return response
			return &response, nil
		})
}

func getBankWithdrawAccount(ctx context.Context, currency string) (model.AccountID, error) {
	bankUser := common.BankUserFromContext(ctx)
	if bankUser.ID == 0 {
		return 0, database.ErrInvalidUserID
	}

	db := appcontext.Database(ctx)
	currencyName := model.CurrencyName(currency)
	if !database.AccountsExists(db, bankUser.ID, currencyName, BankWitdrawAccountName) {
		result, err := AccountCreate(ctx, uint64(bankUser.ID), common.AccountInfo{
			UserID: uint64(bankUser.ID),
			Name:   string(BankWitdrawAccountName),
			Currency: common.CurrencyInfo{
				Name: currency,
			},
		})
		if err != nil {
			return 0, err
		}

		_, err = AccountSetStatus(ctx, result.AccountID, model.AccountStatusNormal.String())
		if err != nil {
			return 0, err
		}
		return model.AccountID(result.AccountID), err
	}

	accounts, err := database.GetAccountsByUserAndCurrencyAndName(db, bankUser.ID, model.CurrencyName(currencyName), BankWitdrawAccountName)
	if err != nil {
		return 0, err
	}

	if len(accounts) == 0 {
		return 0, database.ErrAccountNotFound
	}
	account := accounts[0]
	if account.ID == 0 {
		return 0, database.ErrInvalidAccountID
	}

	return account.ID, nil
}
