/*
 * Copyright 2024 Nxenv, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useState } from 'react'
import { Button, ButtonVariation, Container, Formik, FormikForm, Layout, Text, useToaster } from '@nxenvio/uicore'
import { useHistory } from 'react-router-dom'
import { FontVariation } from '@nxenvio/design-system'
import { useCreateGitspace, type OpenapiCreateGitspaceRequest } from 'cde-rapidship/services'
import RepositoryTypeButton, {
  RepositoryType
} from 'cde-rapidship/components/RepositoryTypeButton/RepositoryTypeButton'
import { useGetSpaceParam } from 'hooks/useGetSpaceParam'
import { useStrings } from 'framework/strings'
import { useAppContext } from 'AppContext'
import { getErrorMessage } from 'utils/Utils'
import { RapidshipRepoImportForm } from 'cde-rapidship/components/RapidshipRepoImportForm/RapidshipRepoImportForm'
import { ThirdPartyRepoImportForm } from 'cde-rapidship/components/ThirdPartyRepoImportForm/ThirdPartyRepoImportForm'
import { CDEIDESelect } from 'cde-rapidship/components/CDEIDESelect/CDEIDESelect'
import { rapidshipFormInitialValues } from './GitspaceCreate.constants'
import { validateRapidshipForm } from './GitspaceCreate.utils'
import css from './GitspaceCreate.module.scss'

export const RapidshipCreateGitspace = () => {
  const { getString } = useStrings()
  const { routes } = useAppContext()
  const history = useHistory()
  const space = useGetSpaceParam()
  const [activeButton, setActiveButton] = useState(RepositoryType.RAPIDSIP)
  const { showSuccess, showError } = useToaster()
  const { mutate } = useCreateGitspace({})

  return (
    <Formik
      onSubmit={async data => {
        try {
          const payload = data
          const response = await mutate({
            ...payload,
            space_ref: space
          } as OpenapiCreateGitspaceRequest & {
            space_ref?: string
          })
          showSuccess(getString('cde.create.gitspaceCreateSuccess'))
          history.push(
            `${routes.toCDEGitspaceDetail({
              space,
              gitspaceId: response.identifier || ''
            })}?redirectFrom=login`
          )
        } catch (error) {
          showError(getString('cde.create.gitspaceCreateFailed'))
          showError(getErrorMessage(error))
        }
      }}
      initialValues={rapidshipFormInitialValues}
      validationSchema={validateRapidshipForm(getString)}
      formName="importRepoForm"
      enableReinitialize>
      {formik => {
        return (
          <>
            <Layout.Horizontal
              className={css.formTitleContainer}
              flex={{ justifyContent: 'space-between', alignItems: 'center' }}>
              <Text font={{ variation: FontVariation.CARD_TITLE }}>{getString('cde.create.repositoryDetails')}</Text>
              <RepositoryTypeButton
                hasChange={formik.dirty}
                onChange={type => {
                  setActiveButton(type)
                  formik?.resetForm()
                }}
              />
            </Layout.Horizontal>
            <FormikForm>
              <Container className={css.formContainer}>
                {activeButton === RepositoryType.RAPIDSIP && (
                  <Container>
                    <RapidshipRepoImportForm />
                  </Container>
                )}
                {activeButton === RepositoryType.THIRDPARTY && (
                  <Container>
                    <ThirdPartyRepoImportForm />
                  </Container>
                )}
              </Container>
              <Container className={css.formOuterContainer}>
                <CDEIDESelect onChange={formik.setFieldValue} selectedIde={formik.values.ide} />
                <Button width={'100%'} variation={ButtonVariation.PRIMARY} height={50} type="submit">
                  {getString('cde.createGitspace')}
                </Button>
              </Container>
            </FormikForm>
          </>
        )
      }}
    </Formik>
  )
}
