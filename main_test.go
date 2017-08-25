package main_test

import (
	"io"
	"net/http"
	"os/exec"
	"strconv"

	"encoding/json"
	"io/ioutil"

	"fmt"

	"os"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/brokerapi"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type failRunner struct {
	Command           *exec.Cmd
	Name              string
	AnsiColorCode     string
	StartCheck        string
	StartCheckTimeout time.Duration
	Cleanup           func()
	session           *gexec.Session
	sessionReady      chan struct{}
}

func (r failRunner) Run(sigChan <-chan os.Signal, ready chan<- struct{}) error {
	defer GinkgoRecover()

	allOutput := gbytes.NewBuffer()

	debugWriter := gexec.NewPrefixedWriter(
		fmt.Sprintf("\x1b[32m[d]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
		GinkgoWriter,
	)

	session, err := gexec.Start(
		r.Command,
		gexec.NewPrefixedWriter(
			fmt.Sprintf("\x1b[32m[o]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
			io.MultiWriter(allOutput, GinkgoWriter),
		),
		gexec.NewPrefixedWriter(
			fmt.Sprintf("\x1b[91m[e]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
			io.MultiWriter(allOutput, GinkgoWriter),
		),
	)

	Ω(err).ShouldNot(HaveOccurred())

	fmt.Fprintf(debugWriter, "spawned %s (pid: %d)\n", r.Command.Path, r.Command.Process.Pid)

	r.session = session
	if r.sessionReady != nil {
		close(r.sessionReady)
	}

	startCheckDuration := r.StartCheckTimeout
	if startCheckDuration == 0 {
		startCheckDuration = 5 * time.Second
	}

	var startCheckTimeout <-chan time.Time
	if r.StartCheck != "" {
		startCheckTimeout = time.After(startCheckDuration)
	}

	detectStartCheck := allOutput.Detect(r.StartCheck)

	for {
		select {
		case <-detectStartCheck: // works even with empty string
			allOutput.CancelDetects()
			startCheckTimeout = nil
			detectStartCheck = nil
			close(ready)

		case <-startCheckTimeout:
			// clean up hanging process
			session.Kill().Wait()

			// fail to start
			return fmt.Errorf(
				"did not see %s in command's output within %s. full output:\n\n%s",
				r.StartCheck,
				startCheckDuration,
				string(allOutput.Contents()),
			)

		case signal := <-sigChan:
			session.Signal(signal)

		case <-session.Exited:
			if r.Cleanup != nil {
				r.Cleanup()
			}

			Expect(string(allOutput.Contents())).To(ContainSubstring(r.StartCheck))
			Expect(session.ExitCode()).To(Not(Equal(0)), fmt.Sprintf("Expected process to exit with non-zero, got: 0"))
			return nil
		}
	}
}

var _ = Describe("Main", func() {
	Context("Missing required args", func() {
		var process ifrit.Process
		It("shows usage", func() {
			var args []string
			blockchainRunner := failRunner{
				Name:       "azureblockchainbroker",
				Command:    exec.Command(binaryPath, args...),
				StartCheck: "Error: adminPassword is required",
			}
			process = ifrit.Invoke(blockchainRunner)
		})

		AfterEach(func() {
			ginkgomon.Kill(process)
		})
	})

	Context("Has required args", func() {
		var (
			args                                                                                                                                                                                                []string
			listenAddr                                                                                                                                                                                          string
			serviceName, serviceID                                                                                                                                                                              string
			username, password                                                                                                                                                                                  string
			tenantID, clientID, clientSecret                                                                                                                                                                    string
			subscriptionID, location, resourceGroupName                                                                                                                                                         string
			namePrefix, adminUsername, adminPassword, ethereumAccountPsswd, ethereumAccountPassphrase, ethereumNetworkID, numConsortiumMembers, numMiningNodesPerMember, mnNodeVMSize, numTXNodes, txNodeVMSize string

			process ifrit.Process
		)

		BeforeEach(func() {
			listenAddr = "127.0.0.1:" + strconv.Itoa(9000+GinkgoParallelNode())
			serviceName = "serviceName"
			serviceID = "serviceID"
			tenantID = "tenantID"
			clientID = "clientID"
			clientSecret = "clientSecret"
			subscriptionID = "subscriptionID"
			location = "location"
			resourceGroupName = "resourceGroupName"
			namePrefix = "namePr"
			adminUsername = "adminUsername"
			adminPassword = "aZure1234567"
			ethereumAccountPsswd = "aZure1234567"
			ethereumAccountPassphrase = "aZure1234567"
			ethereumNetworkID = "123456"
			numConsortiumMembers = "2"
			numMiningNodesPerMember = "1"
			mnNodeVMSize = "Standard_A1"
			numTXNodes = "1"
			txNodeVMSize = "Standard_A1"

			args = append(args, "--listenAddr", listenAddr)
			args = append(args, "--serviceName", serviceName)
			args = append(args, "--serviceID", serviceID)
			args = append(args, "--tenantID", tenantID)
			args = append(args, "--clientID", clientID)
			args = append(args, "--clientSecret", clientSecret)
			args = append(args, "--subscriptionID", subscriptionID)
			args = append(args, "--location", location)
			args = append(args, "--namePrefix", namePrefix)
			args = append(args, "--adminUsername", adminUsername)
			args = append(args, "--adminPassword", adminPassword)
			args = append(args, "--ethereumAccountPsswd", ethereumAccountPsswd)
			args = append(args, "--ethereumAccountPassphrase", ethereumAccountPassphrase)
			args = append(args, "--ethereumNetworkID", ethereumNetworkID)
			args = append(args, "--numConsortiumMembers", numConsortiumMembers)
			args = append(args, "--numMiningNodesPerMember", numMiningNodesPerMember)
			args = append(args, "--mnNodeVMSize", mnNodeVMSize)
			args = append(args, "--numTXNodes", numTXNodes)
			args = append(args, "--txNodeVMSize", txNodeVMSize)

			os.Setenv("USERNAME", username)
			os.Setenv("PASSWORD", password)

		})

		JustBeforeEach(func() {
			blockchainRunner := ginkgomon.New(ginkgomon.Config{
				Name:       "AzureBlockchainBroker",
				Command:    exec.Command(binaryPath, args...),
				StartCheck: "started",
			})
			process = ginkgomon.Invoke(blockchainRunner)
		})

		AfterEach(func() {
			ginkgomon.Kill(process)
		})

		httpDoWithAuth := func(method, endpoint string, body io.ReadCloser) (*http.Response, error) {
			req, err := http.NewRequest(method, "http://"+username+":"+password+"@"+listenAddr+endpoint, body)
			Expect(err).NotTo(HaveOccurred())
			return http.DefaultClient.Do(req)
		}

		It("should listen on the given address", func() {
			resp, err := httpDoWithAuth("GET", "/v2/catalog", nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))

			bytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			var catalog brokerapi.CatalogResponse
			err = json.Unmarshal(bytes, &catalog)
			Expect(err).NotTo(HaveOccurred())

			Expect(catalog.Services[0].Name).To(Equal("serviceName"))
			Expect(catalog.Services[0].ID).To(Equal("serviceID"))
			Expect(catalog.Services[0].Plans[0].ID).To(Equal("7c0b2254-7e68-11e7-bbe1-000d3a818256"))
			Expect(catalog.Services[0].Plans[0].Name).To(Equal("AzureBlockchain"))
			Expect(catalog.Services[0].Plans[0].Description).To(Equal("Azure Blockchain"))
		})
	})
})
