import React, { useState } from 'react'
import { Container, Text } from '@nxenvio/uicore'
import { useStrings } from 'framework/strings'
import { useConfirmAct } from 'hooks/useConfirmAction'
import css from './RepositoryTypeButton.module.scss'

export enum RepositoryType {
  RAPIDSIP = 'rapidship',
  THIRDPARTY = 'thirdParty'
}

const RepositoryTypeButton = ({
  hasChange,
  onChange
}: {
  hasChange?: boolean
  onChange: (type: RepositoryType) => void
}) => {
  const { getString } = useStrings()
  const confirmSwitch = useConfirmAct()
  const [activeButton, setActiveButton] = useState(RepositoryType.RAPIDSIP)

  const handleSwitch = (type: RepositoryType) => {
    const onConfirm = () => {
      setActiveButton(type)
      onChange(type)
    }
    if (hasChange) {
      confirmSwitch({
        title: getString('cde.create.unsaved.title'),
        message: getString('cde.create.unsaved.message'),
        action: async () => {
          onConfirm()
        }
      })
    } else {
      onConfirm()
    }
  }

  return (
    <Container className={css.splitButton}>
      <Text
        className={activeButton === RepositoryType.RAPIDSIP ? css.active : ''}
        onClick={() => {
          handleSwitch(RepositoryType.RAPIDSIP)
        }}>
        {getString('cde.create.rapidshipRepositories')}
      </Text>
      <Text
        className={activeButton === RepositoryType.THIRDPARTY ? css.active : ''}
        onClick={() => {
          handleSwitch(RepositoryType.THIRDPARTY)
        }}>
        {getString('cde.create.thirdPartyGitRepositories')}
      </Text>
    </Container>
  )
}

export default RepositoryTypeButton
