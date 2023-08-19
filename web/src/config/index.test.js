import config from '.';
import { isProduction } from '.';

test('config environment', () => {
  expect(config.environment).toBe('test');
  expect(isProduction()).toBe(false);
});