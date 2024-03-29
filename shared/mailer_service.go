package shared

import (
	"fmt"
	"github.com/JackMaarek/Go-release-test/mailer"
	"github.com/JackMaarek/Go-release-test/shared/repositories"
	"log"
)

const (
	PHISHING_CAMPAIGN_FINISHED_MANAGER_NOTIFICATION_TEMPLATE_ID = "PHISHING_CAMPAIGN_FINISHED_MANAGER_NOTIFICATION"
	PHISHING_CAMPAIGN_FINISHED_EXPERT_NOTIFICATION_TEMPLATE_ID  = "PHISHING_CAMPAIGN_FINISHED_EXPERT_NOTIFICATION"
)

// MailerService handles email operations on phishing.
type MailerService struct {
	SBDetails      *mailer.SBDetails
	Repository 	   *repositories.Repository
}

// SendCompletedCampaignMail send a different email depending on campaign initiator role.
func (ms *MailerService) SendCompletedCampaignMail(companyUUID string, initiatorUUID string) error {
	userRoleList, err := ms.Repository.FindRolesByUserUUID(initiatorUUID)
	if err != nil {
		log.Printf("failed fetching campaign initiator role list from DB with user UUId id %s, reason: %s\n", initiatorUUID, err.Error())
		return fmt.Errorf("failed fetching campaign initiator role list from DB with user UUId id %s, reason: %s\n", initiatorUUID, err.Error())
	}
	isManager := ms.Repository.CheckCompanyManagerRole(userRoleList, repositories.ROLE_COMPANY_MANAGER)
	if isManager {
		err := ms.sendMailToCompanyManager(companyUUID)
		if err != nil {
			fmt.Println(err)
			return err
		}
	} else {
		err := ms.sendMailToInitiator(initiatorUUID)
		if err != nil {
			return err
		}
	}

	return nil
}

// sendMailToCompanyManager takes the companyUUID to gather the managers and create the Address list for the email message.
func (ms *MailerService) sendMailToCompanyManager(companyUUID string) error {
	var user *repositories.User
	var addressList []*mailer.Address
	var sendErrorCount int
	userList, err := ms.Repository.FindUsersByCompanyUUIDAndRoleName(companyUUID, repositories.ROLE_COMPANY_MANAGER)
	if err != nil {
		log.Printf("failed fetching user list from DB with company UUId id %s, reason: %s\n", companyUUID, err.Error())
		return fmt.Errorf("failed fetching user list from DB with company UUId id %s, reason: %s\n", companyUUID, err.Error())
	}
	mt, err := ms.Repository.FindTemplateByInternalID(PHISHING_CAMPAIGN_FINISHED_MANAGER_NOTIFICATION_TEMPLATE_ID)
	if err != nil {
		log.Printf("failed fetching ManagedEmailTemplate object from DB with internal id %s, reason: %s\n",
			PHISHING_CAMPAIGN_FINISHED_MANAGER_NOTIFICATION_TEMPLATE_ID,
			err.Error(),
		)
		return fmt.Errorf("failed fetching ManagedEmailTemplate object from DB with internal id %s, reason: %s\n",
			PHISHING_CAMPAIGN_FINISHED_MANAGER_NOTIFICATION_TEMPLATE_ID,
			err.Error(),
			)
	}
	sendErrorCount = 0
	for _, user = range userList {
		address := mailer.Address{
			Name:  user.Firstname + " " + user.Name,
			Email: user.Email,
		}
		addressList = append(addressList, &address)
		if err != nil {

		}
		err = ms.SBDetails.Send(ms.createMessage(addressList, mt.ProviderID))
		addressList = []*mailer.Address{}
		if err != nil {
			sendErrorCount++
		}
	}
	if sendErrorCount > 0 {
		log.Printf("failed sending %d emails to managers of company with UUId id %s, reason: %s\n", sendErrorCount, companyUUID)
		return fmt.Errorf("failed sending %d emails to managers of company with UUId id %s, reason: %s\n", sendErrorCount, companyUUID)
	}
	return nil
}

// sendMailToInitiator takes the initiatorUUID to gather its information and create the Address for the email message.
func (ms *MailerService) sendMailToInitiator(initiatorUUID string) error {
	var addressList []*mailer.Address
	user, err := ms.Repository.FindInitiatorInformationByUserUUID(initiatorUUID)
	if err != nil {
		log.Printf("failed fetching user information from DB with user UUId id %s, reason: %s\n", initiatorUUID, err.Error())
		return fmt.Errorf("failed fetching user information from DB with user UUId id %s, reason: %s\n", initiatorUUID, err.Error())
	}
	address := mailer.Address{
		Name:  user.Firstname + " " + user.Name,
		Email: user.Email,
	}
	addressList = append(addressList, &address)
	mt, err := ms.Repository.FindTemplateByInternalID(PHISHING_CAMPAIGN_FINISHED_EXPERT_NOTIFICATION_TEMPLATE_ID)
	if err != nil {
		log.Printf("failed fetching ManagedEmailTemplate object from DB with internal id %s, reason: %s\n",
			PHISHING_CAMPAIGN_FINISHED_EXPERT_NOTIFICATION_TEMPLATE_ID,
			err.Error(),
		)
		return fmt.Errorf("failed fetching ManagedEmailTemplate object from DB with internal id %s, reason: %s\n",
			PHISHING_CAMPAIGN_FINISHED_EXPERT_NOTIFICATION_TEMPLATE_ID,
			err.Error(),
		)
	}

	err = ms.SBDetails.Send(ms.createMessage(addressList, mt.ProviderID))
	if err != nil {
		log.Printf("failed sending email to user with UUId id %s, reason: %s\n", initiatorUUID, err.Error())
		return fmt.Errorf("failed sending email to user with UUId id %s, reason: %s\n", initiatorUUID, err.Error())
	}
	return nil
}

// createMessage craft the email Message.
func (ms *MailerService) createMessage(addressList []*mailer.Address, templateId int64) *mailer.Message {
	params := make(map[string]string)
	params["URL"] = "https://app.riskandme.com"
	message := ms.SBDetails.CreateEmailMessage(addressList, templateId, params)

	return message
}
