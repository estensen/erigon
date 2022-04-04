package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon/cmd/observer/database"
	"github.com/ledgerwatch/erigon/cmd/observer/observer"
	"github.com/ledgerwatch/erigon/cmd/observer/reports"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/log/v3"
	"path/filepath"
	"time"
)

func mainWithFlags(ctx context.Context, flags observer.CommandFlags) error {
	server, err := observer.NewServer(flags)
	if err != nil {
		return err
	}

	db, err := database.NewDBSQLite(filepath.Join(flags.DataDir, "observer.sqlite"))
	if err != nil {
		return err
	}

	discV4, err := server.Listen(ctx)
	if err != nil {
		return err
	}

	go observer.StatusLoggerLoop(ctx, db, flags.StatusLogPeriod, log.Root())

	// how many times to retry PING before abandoning a candidate
	const maxPingTries uint = 3
	// the client ID doesn't need to be refreshed often
	const handshakeRefreshTimeout = 20 * 24 * time.Hour
	// how many times to retry handshake before abandoning a candidate
	const maxHandshakeTries uint = 3

	crawlerConfig := observer.CrawlerConfig{
		Chain:            flags.Chain,
		Bootnodes:        server.Bootnodes(),
		PrivateKey:       server.PrivateKey(),
		ConcurrencyLimit: flags.CrawlerConcurrency,
		RefreshTimeout:   flags.RefreshTimeout,
		MaxPingTries:     maxPingTries,
		StatusLogPeriod:  flags.StatusLogPeriod,

		HandshakeRefreshTimeout: handshakeRefreshTimeout,
		MaxHandshakeTries:       maxHandshakeTries,

		KeygenTimeout:     flags.KeygenTimeout,
		KeygenConcurrency: flags.KeygenConcurrency,
	}

	crawler, err := observer.NewCrawler(discV4, db, crawlerConfig, log.Root())
	if err != nil {
		return err
	}

	return crawler.Run(ctx)
}

func reportWithFlags(ctx context.Context, flags reports.CommandFlags) error {
	db, err := database.NewDBSQLite(filepath.Join(flags.DataDir, "observer.sqlite"))
	if err != nil {
		return err
	}

	statusReport, err := reports.CreateStatusReport(ctx, db)
	if err != nil {
		return err
	}
	clientsReport, err := reports.CreateClientsReport(ctx, db, flags.ClientsLimit)
	if err != nil {
		return err
	}

	fmt.Println(statusReport)
	fmt.Println(clientsReport)
	return nil
}

func main() {
	ctx, cancel := common.RootContext()
	defer cancel()

	command := observer.NewCommand()

	reportCommand := reports.NewCommand()
	reportCommand.OnRun(reportWithFlags)
	command.AddSubCommand(reportCommand.RawCommand())

	err := command.ExecuteContext(ctx, mainWithFlags)
	if (err != nil) && !errors.Is(err, context.Canceled) {
		utils.Fatalf("%v", err)
	}
}
