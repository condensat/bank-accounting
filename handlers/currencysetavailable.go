// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-accounting/common"

	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query"

	"github.com/sirupsen/logrus"
)

func CurrencySetAvailable(ctx context.Context, currencyName string, available bool) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CurrencySetAvailable")
	var result common.CurrencyInfo

	log = log.WithField("CurrencyName", currencyName)

	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db database.Context) error {

		// check if currency exists
		currency, err := query.GetCurrencyByName(db, model.CurrencyName(currencyName))
		if err != nil {
			log.WithError(err).Error("Failed to GetCurrencyByName")
			return err
		}

		if string(currency.Name) != currencyName {
			return query.ErrCurrencyNotFound
		}

		if currency.IsAvailable() == available {
			// NOOP
			result = common.CurrencyInfo{
				Name:             string(currency.Name),
				DisplayName:      string(currency.DisplayName),
				Available:        currency.IsAvailable(),
				AutoCreate:       currency.AutoCreate,
				Crypto:           currency.IsCrypto(),
				Type:             common.CurrencyType(currency.GetType()),
				Asset:            currency.GetType() == 2,
				DisplayPrecision: uint(currency.DisplayPrecision()),
			}
			return nil
		}

		var availableState int
		if available {
			availableState = 1
		}

		var crypto model.Int
		if currency.IsCrypto() {
			crypto = 1
		}

		// update currency available
		currency, err = query.AddOrUpdateCurrency(db,
			model.NewCurrency(
				model.CurrencyName(currencyName),
				model.CurrencyName(currency.DisplayName),
				model.Int(currency.GetType()),
				model.Int(availableState),
				crypto,
				currency.DisplayPrecision(),
			),
		)
		if err != nil {
			log.WithError(err).Error("Failed to AddOrUpdateCurrency")
			return err
		}

		result = common.CurrencyInfo{
			Name:             string(currency.Name),
			DisplayName:      string(currency.DisplayName),
			Available:        currency.IsAvailable(),
			AutoCreate:       currency.AutoCreate,
			Crypto:           currency.IsCrypto(),
			Type:             common.CurrencyType(currency.GetType()),
			Asset:            currency.GetType() == 2,
			DisplayPrecision: uint(currency.DisplayPrecision()),
		}

		return nil
	})

	if err == nil {
		log.WithFields(logrus.Fields{
			"Name":        result.Name,
			"DisplayName": result.DisplayName,
			"Available":   result.Available,
			"AutoCreate":  result.AutoCreate,
			"Type":        result.Type,
			"Asset":       result.Asset,
			"Crypto":      result.Crypto,
		}).Warn("Currency updated")
	}

	return result, err
}

func OnCurrencySetAvailable(ctx context.Context, subject string, message *messaging.Message) (*messaging.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Currencying.OnCurrencySetAvailable")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CurrencyInfo
	return messaging.HandleRequest(ctx, appcontext.AppName(ctx), message, &request,
		func(ctx context.Context, _ messaging.BankObject) (messaging.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Name": request.Name,
			})

			currency, err := CurrencySetAvailable(ctx, request.Name, request.Available)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CurrencySetAvailable")
				return nil, cache.ErrInternalError
			}

			log.Info("Currency updated")
			// create & return response
			result := currency
			return &result, nil
		})
}
