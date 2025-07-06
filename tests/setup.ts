import dotenv from 'dotenv';
import { LogLevel, logger } from '../src/utils/logger';

dotenv.config({ path: '.env.test' });

// Set log level to ERROR to reduce noise during tests
logger.setLogLevel(LogLevel.ERROR);