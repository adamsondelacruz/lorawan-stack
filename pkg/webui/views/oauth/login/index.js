// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React from 'react'
import { withRouter } from 'react-router-dom'
import bind from 'autobind-decorator'
import Query from 'query-string'

import api from '../../../api'

import Button from '../../../components/button'
import Field from '../../../components/field'
import Form from '../../../components/form'
import Logo from '../../../components/logo'

import style from './login.styl'

@withRouter
@bind
export default class OAuth extends React.PureComponent {
  constructor (props) {
    super(props)
    this.state = {
      error: '',
    }
  }

  async handleSubmit (values, { setSubmitting, setErrors }) {
    try {
      await api.oauth.login(values)

      window.location = url(this.props.location)
    } catch (error) {
      console.error('yup', error.response)
      this.setState({
        error: error.response.data.message,
      })
    } finally {
      setSubmitting(false)
    }
  }

  render () {

    const initialValues = {
      user_id: '',
      password: '',
    }

    return (
      <div className={style.fullHeightCenter}>
        <div>
          <div className={style.left}>
            <div>
              <Logo />
              <span>Please login to continue</span>
            </div>
          </div>
          <div className={style.right}>
            <h1>The Things Network Account</h1>
            <Form
              onSubmit={this.handleSubmit}
              initialValues={initialValues}
              error={this.state.error}
            >
              <Field
                title="Username or Email"
                name="user_id"
                type="text"
                autoFocus
              />
              <Field
                title="Password"
                name="password"
                type="password"
              />
              <Button type="submit" message="Login" />
              <Button naked message="Create an account" />
            </Form>
          </div>
        </div>
      </div>
    )
  }
}

function url (location, omitQuery = false) {
  const query = Query.parse(location.search)

  const next = query.n || '/'

  if (omitQuery) {
    return next.split('?')[0]
  }

  return next
}