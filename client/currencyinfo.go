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

func CurrencyInfo(ctx context.Context, currencyName string) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CurrencyInfo")

	request := common.CurrencyInfo{
		Name: currencyName,
	}

	var result common.CurrencyInfo
	err := messaging.RequestMessage(ctx, common.CurrencyInfoSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CurrencyInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Name":             result.Name,
		"DisplayName":      result.DisplayName,
		"Available":        result.Available,
		"AutoCreate":       result.AutoCreate,
		"Type":             result.Type,
		"Crypto":           result.Crypto,
		"DisplayPrecision": result.DisplayPrecision,
	}).Trace("Currency CurrencyInfo")

	return result, nil
}
