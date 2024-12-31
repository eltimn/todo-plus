import { sayHello } from './modules/hello.mjs';
import { submitListener } from './modules/forms.mjs'

window.sayHello = sayHello;

// setup login form
// if (window.location.pathname === '/user/login') {

// }

window.Todo = {
  forms: {
    submitListener
  }
}