package broker_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/zeqing-guo/AzureBlockchainBroker/broker"
)

var _ = Describe("AzureConfig", func() {
	var (
		azureConfig *AzureConfig
	)

	Context("Given all required params", func() {
		BeforeEach(func() {
			azureConfig = NewAzureConfig("environment", "tenantID", "clientID", "clientSecret")
		})

		It("should not raise an error", func() {
			err := azureConfig.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Missing environment", func() {
		BeforeEach(func() {
			azureConfig = NewAzureConfig("", "tenantID", "clientID", "clientSecret")
		})

		It("should raise an error", func() {
			err := azureConfig.Validate()
			Expect(err).ToNot(MatchError("Missing required parameters: environment"))
		})
	})

	Context("Missing environment", func() {
		BeforeEach(func() {
			azureConfig = NewAzureConfig("", "tenantID", "clientID", "clientSecret")
		})

		It("should raise an error", func() {
			err := azureConfig.Validate()
			Expect(err).ToNot(MatchError("Missing required parameters: environment"))
		})
	})

	Context("Missing tenantID", func() {
		BeforeEach(func() {
			azureConfig = NewAzureConfig("environment", "", "clientID", "clientSecret")
		})

		It("should raise an error", func() {
			err := azureConfig.Validate()
			Expect(err).ToNot(MatchError("Missing required parameters: tenantID"))
		})
	})

	Context("Missing clientID", func() {
		BeforeEach(func() {
			azureConfig = NewAzureConfig("environment", "tenantID", "", "clientSecret")
		})

		It("should raise an error", func() {
			err := azureConfig.Validate()
			Expect(err).ToNot(MatchError("Missing required parameters: clientID"))
		})
	})

	Context("Missing clientSecret", func() {
		BeforeEach(func() {
			azureConfig = NewAzureConfig("environment", "tenantID", "clientID", "")
		})

		It("should raise an error", func() {
			err := azureConfig.Validate()
			Expect(err).ToNot(MatchError("Missing required parameters: clientSecret"))
		})
	})

	Context("Missing all required params", func() {
		err := azureConfig.Validate()
		Expect(err).ToNot(MatchError("Missing required parameters: environment, tenantID, clientID, clientSecret"))
	})
})

var _ = Describe("AzureStackConfig", func() {
	var (
		azureStackConfig *AzureStackConfig
	)

	Context("Given all required params", func() {
		BeforeEach(func() {
			azureStackConfig = NewAzureStackConfig("azureStackDomain", "azureStackAuthentication", "azureStackResource", "azureStackEndpointPrefix")
		})

		It("should not raise an error", func() {
			err := azureStackConfig.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Missing azureStackDomain", func() {
		BeforeEach(func() {
			azureStackConfig = NewAzureStackConfig("", "azureStackAuthentication", "azureStackResource", "azureStackEndpointPrefix")
		})

		It("should raise an error", func() {
			err := azureStackConfig.Validate()
			Expect(err).To(MatchError("Missing required parameters when 'environment' is 'AzureStack': azureStackDomain"))
		})
	})

	Context("Missing azureStackAuthentication", func() {
		BeforeEach(func() {
			azureStackConfig = NewAzureStackConfig("azureStackDomain", "", "azureStackResource", "azureStackEndpointPrefix")
		})

		It("should raise an error", func() {
			err := azureStackConfig.Validate()
			Expect(err).To(MatchError("Missing required parameters when 'environment' is 'AzureStack': azureStackAuthentication"))
		})
	})

	Context("Missing azureStackResource", func() {
		BeforeEach(func() {
			azureStackConfig = NewAzureStackConfig("azureStackDomain", "azureStackAuthentication", "", "azureStackEndpointPrefix")
		})

		It("should raise an error", func() {
			err := azureStackConfig.Validate()
			Expect(err).To(MatchError("Missing required parameters when 'environment' is 'AzureStack': azureStackResource"))
		})
	})

	Context("Missing azureStackEndpointPrefix", func() {
		BeforeEach(func() {
			azureStackConfig = NewAzureStackConfig("azureStackDomain", "azureStackAuthentication", "azureStackResource", "")
		})

		It("should raise an error", func() {
			err := azureStackConfig.Validate()
			Expect(err).To(MatchError("Missing required parameters when 'environment' is 'AzureStack': azureStackEndpointPrefix"))
		})
	})

	Context("Missing all required params", func() {
		BeforeEach(func() {
			azureStackConfig = NewAzureStackConfig("", "", "", "")
		})

		It("should raise an error", func() {
			err := azureStackConfig.Validate()
			Expect(err).To(MatchError("Missing required parameters when 'environment' is 'AzureStack': azureStackDomain, azureStackAuthentication, azureStackResource, azureStackEndpointPrefix"))
		})
	})
})

var _ = Describe("ResourceConfig", func() {
	var (
		resourceConfig *ResourceConfig
	)

	Context("Given all params", func() {
		BeforeEach(func() {
			resourceConfig = NewResourceConfig("subscriptionID", "resourceGroupName", false, "location", "", false, false)
		})

		It("should not raise an error", func() {
			err := resourceConfig.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Missing subscriptionID", func() {
		BeforeEach(func() {
			resourceConfig = NewResourceConfig("", "resourceGroupName", false, "location", "", false, false)
		})

		It("should raise an error", func() {
			err := resourceConfig.Validate()
			Expect(err).To(MatchError("Missing required parameters: subscriptionID"))
		})
	})

	Context("Missing resourceGroupName", func() {
		BeforeEach(func() {
			resourceConfig = NewResourceConfig("subscriptionID", "", false, "location", "", false, false)
		})

		It("should raise an error", func() {
			err := resourceConfig.Validate()
			Expect(err).To(MatchError("Missing required parameters: resourceGroupName"))
		})
	})

	Context("Missing location", func() {
		BeforeEach(func() {
			resourceConfig = NewResourceConfig("location", "resourceGroupName", false, "location", "", false, false)
		})

		It("should raise an error", func() {
			err := resourceConfig.Validate()
			Expect(err).To(MatchError("Missing required parameters: location"))
		})
	})

	Context("Missing all required params", func() {
		BeforeEach(func() {
			resourceConfig = NewResourceConfig("", "", false, "", "", false, false)
		})

		It("should raise an error", func() {
			err := resourceConfig.Validate()
			Expect(err).To(MatchError("Missing required parameters: subscriptionID, resourceGroupName, location"))
		})
	})
})
