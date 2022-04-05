package observer

import (
	"context"
	"errors"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/internal/debug"
	"github.com/spf13/cobra"
	"github.com/urfave/cli"
	"runtime"
	"time"
)

type CommandFlags struct {
	DataDir            string
	Chain              string
	ListenPort         int
	NATDesc            string
	NetRestrict        string
	NodeKeyFile        string
	NodeKeyHex         string
	Bootnodes          string
	CrawlerConcurrency uint
	RefreshTimeout     time.Duration
	KeygenTimeout      time.Duration
	KeygenConcurrency  uint
	StatusLogPeriod    time.Duration
}

type Command struct {
	command cobra.Command
	flags   CommandFlags
}

func NewCommand() *Command {
	command := cobra.Command{
		Short: "P2P network crawler",
	}

	// debug flags
	utils.CobraFlags(&command, append(debug.Flags, utils.MetricFlags...))

	instance := Command{
		command: command,
	}
	instance.withDatadir()
	instance.withChain()
	instance.withListenPort()
	instance.withNAT()
	instance.withNetRestrict()
	instance.withNodeKeyFile()
	instance.withNodeKeyHex()
	instance.withBootnodes()
	instance.withCrawlerConcurrency()
	instance.withRefreshTimeout()
	instance.withKeygenTimeout()
	instance.withKeygenConcurrency()
	instance.withStatusLogPeriod()

	return &instance
}

func (command *Command) withDatadir() {
	flag := utils.DataDirFlag
	command.command.Flags().StringVar(&command.flags.DataDir, flag.Name, flag.Value.String(), flag.Usage)
	must(command.command.MarkFlagDirname(utils.DataDirFlag.Name))
}

func (command *Command) withChain() {
	flag := utils.ChainFlag
	command.command.Flags().StringVar(&command.flags.Chain, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) withListenPort() {
	flag := utils.ListenPortFlag
	command.command.Flags().IntVar(&command.flags.ListenPort, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) withNAT() {
	flag := utils.NATFlag
	command.command.Flags().StringVar(&command.flags.NATDesc, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) withNetRestrict() {
	flag := utils.NetrestrictFlag
	command.command.Flags().StringVar(&command.flags.NetRestrict, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) withNodeKeyFile() {
	flag := utils.NodeKeyFileFlag
	command.command.Flags().StringVar(&command.flags.NodeKeyFile, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) withNodeKeyHex() {
	flag := utils.NodeKeyHexFlag
	command.command.Flags().StringVar(&command.flags.NodeKeyHex, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) withBootnodes() {
	flag := utils.BootnodesFlag
	command.command.Flags().StringVar(&command.flags.Bootnodes, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) withCrawlerConcurrency() {
	flag := cli.UintFlag{
		Name:  "crawler-concurrency",
		Usage: "A number of maximum parallel node interrogations",
		Value: uint(runtime.GOMAXPROCS(-1)) * 10,
	}
	command.command.Flags().UintVar(&command.flags.CrawlerConcurrency, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) withRefreshTimeout() {
	flag := cli.DurationFlag{
		Name:  "refresh-timeout",
		Usage: "A timeout to wait before considering to re-crawl a node",
		Value: 2 * 24 * time.Hour,
	}
	command.command.Flags().DurationVar(&command.flags.RefreshTimeout, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) withKeygenTimeout() {
	flag := cli.DurationFlag{
		Name:  "keygen-timeout",
		Usage: "How much time can be used to generate node bucket keys",
		Value: 2 * time.Second,
	}
	command.command.Flags().DurationVar(&command.flags.KeygenTimeout, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) withKeygenConcurrency() {
	flag := cli.UintFlag{
		Name:  "keygen-concurrency",
		Usage: "How many parallel goroutines can be used by the node bucket keys generator",
		Value: 2,
	}
	command.command.Flags().UintVar(&command.flags.KeygenConcurrency, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) withStatusLogPeriod() {
	flag := cli.DurationFlag{
		Name:  "status-log-period",
		Usage: "How often to log status summaries",
		Value: 10 * time.Second,
	}
	command.command.Flags().DurationVar(&command.flags.StatusLogPeriod, flag.Name, flag.Value, flag.Usage)
}

func (command *Command) ExecuteContext(ctx context.Context, runFunc func(ctx context.Context, flags CommandFlags) error) error {
	command.command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// apply debug flags
		return utils.SetupCobra(cmd)
	}
	command.command.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		debug.Exit()
	}
	command.command.RunE = func(cmd *cobra.Command, args []string) error {
		defer debug.Exit()
		err := runFunc(cmd.Context(), command.flags)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
	return command.command.ExecuteContext(ctx)
}

func (command *Command) AddSubCommand(subCommand *cobra.Command) {
	command.command.AddCommand(subCommand)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
