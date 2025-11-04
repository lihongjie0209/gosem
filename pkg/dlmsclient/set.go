package dlmsclient

import (
	"errors"
	"fmt"
	"reflect"

	"gitlab.com/circutor-library/gosem/pkg/axdr"
	"gitlab.com/circutor-library/gosem/pkg/dlms"
)

func (c *client) SetRequest(att *dlms.AttributeDescriptor, data interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.setRequest(att, data)
}

func (c *client) SetRequestWithList(att []*dlms.AttributeDescriptor, data []interface{}) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.setRequestWithList(att, data)
}

//nolint:nestif
func (c *client) SetRequestWithStructOfElements(data interface{}, continueOnSetRejected bool) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	v := eindirect(reflect.ValueOf(data))

	if v.Kind() != reflect.Struct {
		return dlms.NewError(dlms.ErrorInvalidParameter, "data must be a struct")
	}

	var errSet error
	isSomethingDone := false
	isSomethingFailed := false

	for i := 0; i < v.NumField(); i++ {
		ad, err := c.getAttributeDescriptor(v.Type().Field(i))
		if err != nil {
			return err
		}

		if ad == nil {
			continue
		}

		if v.Field(i).Kind() == reflect.Pointer {
			// All fields need to have been set beforehand: nil fields will be ignored
			if v.Field(i).IsNil() {
				continue
			}
		}

		err = c.setRequest(ad, v.Field(i).Interface())
		if err != nil {
			// If a set is rejected, we will continue anyway
			var dlmsError *dlms.Error
			if errors.As(err, &dlmsError) && dlmsError.Code() == dlms.ErrorSetRejected && continueOnSetRejected {
				isSomethingFailed = true
			} else {
				if isSomethingDone {
					err = dlms.NewError(dlms.ErrorSetPartial, fmt.Sprintf("partial set: %v", err))
				}

				return err
			}

			errSet = err
		} else {
			isSomethingDone = true
		}
	}

	if isSomethingFailed && isSomethingDone {
		errSet = dlms.NewError(dlms.ErrorSetPartial, fmt.Sprintf("partial set: %v", errSet))
	}

	return errSet
}

func (c *client) setRequest(att *dlms.AttributeDescriptor, data interface{}) (err error) {
	if att == nil {
		return dlms.NewError(dlms.ErrorInvalidParameter, "attribute descriptor must be non-nil")
	}

	dt, ok := data.(*axdr.DlmsData)
	if !ok {
		dt, err = axdr.MarshalData(data)
		if err != nil {
			return dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error marshaling %s data: %v", att.String(), err))
		}
	}

	out, err := dt.Encode()
	if err != nil {
		return dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error encoding %s data: %v", att.String(), err))
	}

	lenHeader := 13
	if c.settings.Ciphering.Level != dlms.SecurityLevelNone {
		lenHeader = 34
	}

	if len(out) < (c.settings.MaxPduSendSize - lenHeader) {
		req := dlms.CreateSetRequestNormal(unicastInvokeID, *att, nil, *dt)

		pdu, err := c.encodeSendReceiveAndDecode(req)
		if err != nil {
			return err
		}

		resp, ok := pdu.(dlms.SetResponseNormal)
		if !ok {
			return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in %s unexpected PDU response type: %T", att.String(), pdu))
		}

		if resp.Result != dlms.TagAccSuccess {
			return dlms.NewError(dlms.ErrorSetRejected, fmt.Sprintf("set %s rejected: %s", att.String(), resp.Result.String()))
		}
	} else {
		return c.setRequestWithDataBlock(att, out)
	}

	return
}

func (c *client) setRequestWithDataBlock(att *dlms.AttributeDescriptor, out []byte) error {
	isLastBlock := false
	isFirstBlock := true
	blockNumber := uint32(1)

	for {
		lenHeader := 11
		if isFirstBlock {
			lenHeader = 21
		}
		if c.settings.Ciphering.Level != dlms.SecurityLevelNone {
			lenHeader += 21
		}

		blockSize := c.settings.MaxPduSendSize - lenHeader
		if blockSize > len(out) {
			blockSize = len(out)
			isLastBlock = true
		}

		db := dlms.CreateDataBlockSA(isLastBlock, blockNumber, out[:blockSize])

		var req dlms.CosemPDU

		if isFirstBlock {
			req = dlms.CreateSetRequestWithFirstDataBlock(unicastInvokeID, *att, nil, *db)
		} else {
			req = dlms.CreateSetRequestWithDataBlock(unicastInvokeID, *db)
		}

		pdu, err := c.encodeSendReceiveAndDecode(req)
		if err != nil {
			return err
		}

		if isLastBlock {
			resp, ok := pdu.(dlms.SetResponseLastDataBlock)
			if !ok {
				return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in %s unexpected PDU response type: %T", att.String(), pdu))
			}

			if resp.BlockNum != blockNumber {
				return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in %s unexpected block number %d (expected %d)", att.String(), resp.BlockNum, blockNumber))
			}

			if resp.Result != dlms.TagAccSuccess {
				return dlms.NewError(dlms.ErrorSetRejected, fmt.Sprintf("set %s rejected: %s", att.String(), resp.Result.String()))
			}

			return nil
		}

		resp, ok := pdu.(dlms.SetResponseDataBlock)
		if !ok {
			return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in %s unexpected PDU response type: %T", att.String(), pdu))
		}

		if resp.BlockNum != blockNumber {
			return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in %s unexpected block number %d (expected %d)", att.String(), resp.BlockNum, blockNumber))
		}

		isFirstBlock = false
		out = out[blockSize:]
		blockNumber++
	}
}

func (c *client) setRequestWithList(att []*dlms.AttributeDescriptor, data []interface{}) (err error) {
	if len(att) != len(data) {
		return dlms.NewError(dlms.ErrorInvalidParameter, "attribute descriptor list and data list must have the same length")
	}

	output := []byte{byte(len(data))}

	for i := range data {
		dt, ok := data[i].(*axdr.DlmsData)
		if !ok {
			dt, err = axdr.MarshalData(data[i])
			if err != nil {
				return dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error marshaling %s data: %v", att[i].String(), err))
			}
		}

		out, err := dt.Encode()
		if err != nil {
			return dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error encoding %s data: %v", att[i].String(), err))
		}

		output = append(output, out...)
	}

	// Currently only supports sending WithListAndBlock
	return c.setRequestWithListAndDataBlock(att, output)
}

func (c *client) setRequestWithListAndDataBlock(att []*dlms.AttributeDescriptor, out []byte) error {
	isLastBlock := false
	isFirstBlock := true
	blockNumber := uint32(1)

	for {
		lenHeader := 11
		if isFirstBlock {
			lenHeader += len(att) * 10
		}
		if c.settings.Ciphering.Level != dlms.SecurityLevelNone {
			lenHeader += 21
		}

		blockSize := c.settings.MaxPduSendSize - lenHeader
		if blockSize > len(out) {
			blockSize = len(out)
			isLastBlock = true
		}

		db := dlms.CreateDataBlockSA(isLastBlock, blockNumber, out[:blockSize])

		var req dlms.CosemPDU

		if isFirstBlock {
			attList := make([]dlms.AttributeDescriptorWithSelection, len(att))
			for i := range att {
				attList[i] = dlms.AttributeDescriptorWithSelection{
					ClassID:          att[i].ClassID,
					InstanceID:       att[i].InstanceID,
					AttributeID:      att[i].AttributeID,
					AccessDescriptor: nil,
				}
			}

			req = dlms.CreateSetRequestWithListAndFirstDataBlock(unicastInvokeID, attList, *db)
		} else {
			req = dlms.CreateSetRequestWithDataBlock(unicastInvokeID, *db)
		}

		pdu, err := c.encodeSendReceiveAndDecode(req)
		if err != nil {
			return err
		}

		if isLastBlock {
			resp, ok := pdu.(dlms.SetResponseLastDataBlockWithList)
			if !ok {
				return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in set with list, unexpected PDU response type: %T", pdu))
			}

			if resp.BlockNum != blockNumber {
				return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in set with list, unexpected block number %d (expected %d)", resp.BlockNum, blockNumber))
			}

			for i, respTag := range resp.ResultList {
				if respTag != dlms.TagAccSuccess {
					return dlms.NewError(dlms.ErrorSetRejected, fmt.Sprintf("set %s rejected: %s", att[i].String(), respTag.String()))
				}
			}

			return nil
		}

		resp, ok := pdu.(dlms.SetResponseDataBlock)
		if !ok {
			return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in set with list, unexpected PDU response type: %T", pdu))
		}

		if resp.BlockNum != blockNumber {
			return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("in set with list, unexpected block number %d (expected %d)", resp.BlockNum, blockNumber))
		}

		isFirstBlock = false
		out = out[blockSize:]
		blockNumber++
	}
}

func eindirect(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return eindirect(v.Elem())
	default:
		return v
	}
}
