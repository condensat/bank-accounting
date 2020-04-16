// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"

	"github.com/condensat/bank-accounting/common"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func CurrencyCreate(ctx context.Context, currencyName string, isCrypto bool, displayPrecision uint) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CurrencyCreate")

	request := common.CurrencyInfo{
		Name:             currencyName,
		Crypto:           isCrypto,
		DisplayPrecision: displayPrecision,
	}

	var result common.CurrencyInfo
	err := messaging.RequestMessage(ctx, common.CurrencyCreateSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CurrencyInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Name":             result.Name,
		"Available":        result.Available,
		"Crypto":           result.Crypto,
		"DisplayPrecision": result.DisplayPrecision,
	}).Debug("Currency Created")

	return result, nil
}