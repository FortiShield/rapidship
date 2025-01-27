import React from 'react'
import { Container } from '@nxenvio/uicore'
import type { UsefulOrNotProps } from 'utils/types'

export const defaultUsefulOrNot = (props: UsefulOrNotProps): React.ReactElement => {
  return <Container {...props} />
}
