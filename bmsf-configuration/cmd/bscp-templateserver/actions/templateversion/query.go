/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package templateversion

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// QueryAction query target config template version object.
type QueryAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryTemplateVersionReq
	resp *pb.QueryTemplateVersionResp

	templateVersion *pbcommon.TemplateVersion
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryTemplateVersionReq, resp *pb.QueryTemplateVersionResp) *QueryAction {
	action := &QueryAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *QueryAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *QueryAction) Output() error {
	act.resp.TemplateVersion = act.templateVersion
	return nil
}

func (act *QueryAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Versionid, "versionid"); err != nil {
		return err
	}

	return nil
}

func (act *QueryAction) queryTemplateVersion() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryTemplateVersionReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Versionid: act.req.Versionid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryTemplateVersion[%d]| request to datamanager QueryTemplateVersion, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryTemplateVersion(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryTemplateVersion, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	act.templateVersion = resp.TemplateVersion

	return pbcommon.ErrCode_E_OK, ""
}

// Do do action.
func (act *QueryAction) Do() error {
	// query config template version.
	if errCode, errMsg := act.queryTemplateVersion(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
